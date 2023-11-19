package book

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/database/book"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/tools"

	"github.com/gin-gonic/gin"
)

func GetBooks(c *gin.Context) {
	var (
		response []models.QueryBooks
	)
	db := database.GetDB()
	r := book.NewRepository(db)

	queryList := c.QueryArray("text")
	limit, _ := strconv.Atoi(c.Query("count"))
	uniqueQueryList := tools.UniqueSlice(queryList)
	response, errors := collectBooksFromHistory(r, uniqueQueryList, limit)
	fmt.Println(errors)

	c.JSON(200, response)
}

func collectBooksFromHistory(r *book.Repository, queryList []string, limit int) ([]models.QueryBooks, []error) {
	var (
		response          []models.QueryBooks
		notFoundedQueries []string
		selectErrors      []error
		wg                sync.WaitGroup
	)
	wg.Add(len(queryList))

	for _, query := range queryList {
		go func(query string) {
			defer wg.Done()

			books, err := r.GetByQuery(query, limit)
			if err != nil {
				selectErrors = append(selectErrors, err)
				return
			}
			if len(books) > 0 {
				response = append(response, models.QueryBooks{
					Query:    query,
					BookList: books,
				})
			} else {
				fmt.Println("не найден:", query)
				notFoundedQueries = append(notFoundedQueries, query)
			}

		}(query)
	}
	wg.Wait()

	books, errors := collectNotFoundedBooks(r, notFoundedQueries, limit)
	response = append(response, books...)
	selectErrors = append(selectErrors, errors...)
	return response, selectErrors
}

func collectNotFoundedBooks(r *book.Repository, queryList []string, limit int) ([]models.QueryBooks, []error) {
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
			data := models.QueryBooks{
				Query:    query,
				BookList: books,
			}
			response = append(response, data)
			r.SaveBooks(data)

		}(query)
	}
	wg.Wait()

	return response, selectErrors
}
