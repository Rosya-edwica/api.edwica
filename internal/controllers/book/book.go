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

var BookCache = map[string][]models.Book{}
var BookCacheMutex = sync.RWMutex{}

func GetBooks(c *gin.Context) {
	var response []models.QueryBooks
	r := book.NewRepository(database.GetDB())
	queryList, limit := valideBookParams(c)

	cacheResponse, notFounded := checkNewQueriesInCache(tools.UniqueSlice(queryList))
	response, notFounded, _ = collectBooksFromHistory(r, notFounded, limit)
	newBooks, errors := collectNotFoundedBooks(r, notFounded, limit)
	response = append(response, newBooks...)
	errors = append(errors, errors...)
	response = append(response, cacheResponse...)
	c.JSON(200, response)
}

func valideBookParams(c *gin.Context) (queryList []string, limit int) {
	queryList = c.QueryArray("text")
	if len(queryList) == 0 {
		c.JSON(207, controllers.IncorrentParameters(c, "text"))
		return
	}

	limit, err := strconv.Atoi(c.Query("count"))
	if err != nil || limit == 0 {
		limit = DefaultLimit
	}
	return
}

func checkNewQueriesInCache(items []string) (cacheResponse []models.QueryBooks, notFoundedInCache []string) {
	for _, query := range items {
		BookCacheMutex.RLock()
		if val, ok := BookCache[query]; ok {
			cacheResponse = append(cacheResponse, models.QueryBooks{Query: query, BookList: val})
		} else {
			notFoundedInCache = append(notFoundedInCache, query)
		}
		BookCacheMutex.RUnlock()
	}
	return cacheResponse, notFoundedInCache
}

func collectBooksFromHistory(r *book.Repository, queryList []string, limit int) (response []models.QueryBooks, notFounded []string, errors []error) {
	wg := sync.WaitGroup{}
	wg.Add(len(queryList))

	for _, query := range queryList {
		go func(query string) {
			defer wg.Done()

			books, err := r.GetByQuery(query, limit)
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
				BookCacheMutex.Lock()
				BookCache[query] = books
				BookCacheMutex.Unlock()
			} else {
				fmt.Println("не найден:", query)
				notFounded = append(notFounded, query)
			}

		}(query)
	}
	wg.Wait()

	return
}

func collectNotFoundedBooks(r *book.Repository, queryList []string, limit int) ([]models.QueryBooks, []error) {
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

			books, err := r.GetByName(query, limit)
			if err != nil {
				selectErrors = append(selectErrors, err)
				return
			}
			if books == nil {
				books = []models.Book{}
				r.SaveUndefiendQuery(query)
			}
			data := models.QueryBooks{
				Query:    query,
				BookList: books,
			}
			response = append(response, data)
			BookCacheMutex.Lock()
			BookCache[query] = books
			BookCacheMutex.Unlock()
			r.SaveBooks(data)

		}(query)
	}
	wg.Wait()
	logger.Log.Info(fmt.Sprintf("Для этих значений не было данных в истории %#v\n", queryList))
	return response, selectErrors
}
