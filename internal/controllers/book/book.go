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

func GetBooks(c *gin.Context) {
	var (
		response []models.QueryBooks
	)
	db := database.GetDB()
	r := book.NewRepository(db)

	queryList := c.QueryArray("text")
	if len(queryList) == 0 {
		c.JSON(207, controllers.IncorrentParameters(c, "text"))
		return
	}

	limit, err := strconv.Atoi(c.Query("count"))
	if err != nil || limit == 0 {
		limit = DefaultLimit
	}
	uniqueQueryList := tools.UniqueSlice(queryList)
	response, notFounded, _ := collectBooksFromHistory(r, uniqueQueryList, limit)
	// errMap := controllers.ErrorListHandler(c, errors)
	// if errMap != nil {
	// 	c.JSON(404, errMap)
	// 	return
	// }

	newBooks, errors := collectNotFoundedBooks(r, notFounded, limit)
	response = append(response, newBooks...)
	errors = append(errors, errors...)
	c.JSON(200, response)
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
				response = append(response, models.QueryBooks{
					Query:    query,
					BookList: books,
				})
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
			r.SaveBooks(data)

		}(query)
	}
	wg.Wait()
	logger.Log.Info(fmt.Sprintf("Для этих значений не было данных в истории %#v\n", queryList))
	return response, selectErrors
}
