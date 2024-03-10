package book

import (
	"fmt"
	"github.com/go-faster/errors"
	"strconv"
	"sync"

	"github.com/Rosya-edwica/api.edwica/internal/controllers"
	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/database/book"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/Rosya-edwica/api.edwica/pkg/tools"

	"github.com/gin-gonic/gin"
)

const DefaultLimit = 3
const DefaultSubDomain = ""

// Будем хранить такие строки: "text=query&limit=3&subdomain=ssssss"
// Это нужно для того, чтобы в кэше хранилась инфа не только о названии книги, но и количестве с поддоменом
var BookCache = map[string][]models.Book{}
var BookCacheMutex = sync.RWMutex{}

// GetBooks Обработчик /books
func GetBooks(c *gin.Context) {
	var response []models.QueryBooks
	params := valideBookParams(c)

	response, notFounded := checkNewQueriesInCache(params)
	params.Texts = notFounded
	if params.Source == "" && len(params.Texts) != 0 {
		litresResp, _ := findBooksInLitres(params)
		response = append(response, litresResp...)
	} else if params.Source != "" {
		DBResp, _ := findBooksInHistory(params)
		response = append(response, DBResp...)
	}
	c.JSON(200, response)

}

func findBooksInLitres(params models.BookParams) (response []models.QueryBooks, err error) {
	var wg sync.WaitGroup
	wg.Add(len(params.Texts))
	for _, i := range params.Texts {
		go func(text string) {
			defer wg.Done()
			var res models.QueryBooks
			res, err = SearchBooks(text, params.Limit)

			response = append(response, res)
			requestStr := createCacheRequestStr(text, params.Limit, params.Source)
			addRequestToCache(requestStr, res.BookList)
		}(i)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return
}

func findBooksInHistory(params models.BookParams) (result []models.QueryBooks, err error) {
	r := book.NewRepository(database.GetDB())
	response, notFounded, errList := collectBooksFromHistory(r, params)
	if len(errList) != 0 {
		return nil, errors.Wrap(errList[0], "Ошибка при чтении книг из истории запросов")
	}
	params.Texts = notFounded
	newBooks, errList := collectNotFoundedBooks(r, params)
	if len(errList) != 0 {
		return nil, errors.Wrap(errList[0], "Ошибка при поиске книг в БД")
	}
	result = append(result, response...)
	result = append(response, newBooks...)
	if len(result) != 0 {
		fmt.Printf("limit:%d\tcount books:%d\n", params.Limit, len(result[0].BookList))
	}
	return
}

// valideBookParams
func valideBookParams(c *gin.Context) models.BookParams {
	queryList := c.QueryArray("text")
	if len(queryList) == 0 {
		c.JSON(207, controllers.IncorrentParameters(c, "text"))
		return models.BookParams{}
	}
	limit, err := strconv.Atoi(c.Query("count"))
	if err != nil || limit == 0 {
		limit = DefaultLimit
	}
	source := c.Query("domain")
	if len(source) == 0 {
		source = DefaultSubDomain
	}

	return models.BookParams{
		Texts:  tools.UniqueSlice(queryList),
		Limit:  limit,
		Source: source,
	}
}

func checkNewQueriesInCache(params models.BookParams) (cacheResponse []models.QueryBooks, notFoundedInCache []string) {
	for _, query := range params.Texts {
		BookCacheMutex.RLock()
		requestStr := createCacheRequestStr(query, params.Limit, params.Source)
		if val, ok := BookCache[requestStr]; ok {
			fmt.Println("Прочитали из кэша: ", requestStr)
			cacheResponse = append(cacheResponse, models.QueryBooks{Query: query, BookList: val})
		} else {
			notFoundedInCache = append(notFoundedInCache, query)
		}
		BookCacheMutex.RUnlock()
	}
	return cacheResponse, notFoundedInCache
}

func collectBooksFromHistory(r *book.Repository, params models.BookParams) (response []models.QueryBooks, notFounded []string, errors []error) {
	wg := sync.WaitGroup{}
	wg.Add(len(params.Texts))

	for _, query := range params.Texts {
		go func(query string) {
			defer wg.Done()
			books, err := r.GetByQuery(query, params.Limit, params.Source)
			if err != nil {
				errors = append(errors, err)
				return
			}
			if len(books) > 0 {
				item := models.QueryBooks{
					Query:    query,
					BookList: books,
				}
				response = append(response, item)
				requestStr := createCacheRequestStr(query, params.Limit, params.Source)
				addRequestToCache(requestStr, books)
			} else {
				fmt.Println("не найден:", query)
				notFounded = append(notFounded, query)
			}
		}(query)
	}
	wg.Wait()
	return
}

func collectNotFoundedBooks(r *book.Repository, params models.BookParams) ([]models.QueryBooks, []error) {
	if len(params.Texts) == 0 {
		return nil, nil
	}
	var (
		response     []models.QueryBooks
		selectErrors []error
		wg           sync.WaitGroup
	)
	wg.Add(len(params.Texts))

	for _, query := range params.Texts {
		go func(query string) {
			defer wg.Done()
			var data models.QueryBooks

			books, err := r.GetByName(query, params.Source)
			if err != nil {
				selectErrors = append(selectErrors, err)
				return
			}
			if books == nil {
				books = []models.Book{}
				r.SaveUndefiendQuery(query)
			}
			// Возвращаем пользователю limit книг
			if len(books) > params.Limit {
				data = models.QueryBooks{
					Query:    query,
					BookList: books[:params.Limit],
				}
			} else {
				data = models.QueryBooks{
					Query:    query,
					BookList: books,
				}
			}

			response = append(response, data)
			requestStr := createCacheRequestStr(query, params.Limit, params.Source)
			addRequestToCache(requestStr, books)
			// Сохраняем все книги, а не limit штук
			r.SaveBooks(models.QueryBooks{
				Query:    query,
				BookList: books,
			})

		}(query)
	}
	wg.Wait()
	logger.Log.Info(fmt.Sprintf("Для этих значений не было данных в истории %#v\n", params.Texts))
	return response, selectErrors
}

// Что сохраняем в кэш? Ключ - название книги + количество книг +  поддомен. Значение - список из Book
// Зачем? Чтобы лишний раз не лезть в базу данных
func addRequestToCache(requestStr string, books []models.Book) {
	fmt.Println("Записали в кэш: ", requestStr)
	BookCacheMutex.Lock()
	BookCache[requestStr] = books
	BookCacheMutex.Unlock()
}

// Формируем строчку для ключа мапы с кэшем
func createCacheRequestStr(query string, limit int, subdomain string) string {
	return fmt.Sprintf("text=%s&limit=%d&subdomain=%s", query, limit, subdomain)
}
