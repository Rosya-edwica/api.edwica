package book

import (
	"fmt"
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

func GetBooks(c *gin.Context) {
	var response []models.QueryBooks
	r := book.NewRepository(database.GetDB())
	queryList, subdomain, limit := valideBookParams(c)

	cacheResponse, notFounded := checkNewQueriesInCache(tools.UniqueSlice(queryList), limit, subdomain)
	response, notFounded, _ = collectBooksFromHistory(r, notFounded, subdomain, limit)
	newBooks, errors := collectNotFoundedBooks(r, notFounded, subdomain, limit)
	response = append(response, newBooks...)
	errors = append(errors, errors...)
	response = append(response, cacheResponse...)
	fmt.Printf("limit:%d\tcount books:%d\n", limit, len(response[0].BookList))
	c.JSON(200, response)
}

func valideBookParams(c *gin.Context) (queryList []string, subdomain string, limit int) {
	queryList = c.QueryArray("text")
	if len(queryList) == 0 {
		c.JSON(207, controllers.IncorrentParameters(c, "text"))
		return
	}
	limit, err := strconv.Atoi(c.Query("count"))
	if err != nil || limit == 0 {
		limit = DefaultLimit
	}
	subdomain = c.Query("domain")
	if len(subdomain) == 0 {
		subdomain = DefaultSubDomain
	}

	return
}

func checkNewQueriesInCache(items []string, limit int, subdomain string) (cacheResponse []models.QueryBooks, notFoundedInCache []string) {
	for _, query := range items {
		BookCacheMutex.RLock()
		requestStr := createCacheRequestStr(query, limit, subdomain)
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

func collectBooksFromHistory(r *book.Repository, queryList []string, subdomain string, limit int) (response []models.QueryBooks, notFounded []string, errors []error) {
	wg := sync.WaitGroup{}
	wg.Add(len(queryList))

	for _, query := range queryList {
		go func(query string) {
			defer wg.Done()
			books, err := r.GetByQuery(query, limit, subdomain)
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
				requestStr := createCacheRequestStr(query, limit, subdomain)
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

func collectNotFoundedBooks(r *book.Repository, queryList []string, subdomain string, limit int) ([]models.QueryBooks, []error) {
	if len(queryList) == 0 {
		return nil, nil
	}
	var (
		response     []models.QueryBooks
		selectErrors []error
		wg           sync.WaitGroup
	)
	wg.Add(len(queryList))

	for _, query := range queryList {
		go func(query string) {
			defer wg.Done()
			var data models.QueryBooks

			books, err := r.GetByName(query, subdomain)
			if err != nil {
				selectErrors = append(selectErrors, err)
				return
			}
			if books == nil {
				books = []models.Book{}
				r.SaveUndefiendQuery(query)
			}
			// Возвращаем пользователю limit книг
			if len(books) > limit {
				data = models.QueryBooks{
					Query:    query,
					BookList: books[:limit],
				}
			} else {
				data = models.QueryBooks{
					Query:    query,
					BookList: books,
				}
			}

			response = append(response, data)
			requestStr := createCacheRequestStr(query, limit, subdomain)
			addRequestToCache(requestStr, books)
			// Сохраняем все книги, а не limit штук
			r.SaveBooks(models.QueryBooks{
				Query:    query,
				BookList: books,
			})

		}(query)
	}
	wg.Wait()
	logger.Log.Info(fmt.Sprintf("Для этих значений не было данных в истории %#v\n", queryList))
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
