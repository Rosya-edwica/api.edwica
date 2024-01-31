package book

import (
	"testing"

	"github.com/Rosya-edwica/api.edwica/config"
	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/database/book"
	"github.com/stretchr/testify/assert"
)

func getRep() *book.Repository {
	cfg, _ := config.LoadDBConfig("../../../config")
	db, _ := database.New(cfg)
	r := book.NewRepository(db)
	return r
}
func TestCollectBooksFromHistory(t *testing.T) {
	r := getRep()

	var (
		queryList = []string{"Golang", "Python", "Linux", "fdinaiundnaudnuansdunasd"}
		limit     = 5
		subdomain = "svoevagro"
	)

	data, notFounded, errors := collectBooksFromHistory(r, queryList, subdomain, limit)

	for _, err := range errors {
		assert.Nil(t, err)
	}
	assert.Len(t, data, 3)
	assert.NotEmpty(t, notFounded)

	for _, i := range data {
		assert.Contains(t, queryList, i.Query)
		assert.NotEmpty(t, i.BookList)
		assert.LessOrEqual(t, len(i.BookList), limit)
		for _, b := range i.BookList {
			assert.NotEmpty(t, b.Id)
			assert.NotEmpty(t, b.Name)
		}
	}

	err := database.Close(database.GetDB())
	assert.Nil(t, err)
}

func TestCollectNotFoundedBooks(t *testing.T) {
	r := getRep()

	var (
		subdomain = "svoevagro"
		queryList = []string{"Программирование"}
		limit     = 3
	)

	data, errors := collectNotFoundedBooks(r, queryList, subdomain, limit)

	for _, err := range errors {
		assert.Nil(t, err)
	}
	assert.Len(t, data, len(queryList))
	for _, i := range data {
		assert.Contains(t, queryList, i.Query)
		assert.NotEmpty(t, i.BookList)
		assert.LessOrEqual(t, len(i.BookList), limit)
		for _, b := range i.BookList {
			assert.NotEmpty(t, b.Id)
			assert.NotEmpty(t, b.Name)
		}
	}
	err := database.Close(database.GetDB())
	assert.Nil(t, err)
}
