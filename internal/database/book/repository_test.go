package book

import (
	"strings"
	"testing"

	"github.com/Rosya-edwica/api.edwica/config"
	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/stretchr/testify/assert"
)

func getRep() *Repository {
	cfg, _ := config.LoadDBConfig("../../../config")
	db, _ := database.New(cfg)
	r := NewRepository(db)
	return r
}

func TestConnection(t *testing.T) {
	cfg, err := config.LoadDBConfig("../../../config")
	assert.Nil(t, err)

	db, err := database.New(cfg)
	assert.Nil(t, err)
	err = database.Close(db)
	assert.Nil(t, err)
}

func TestGetByName(t *testing.T) {
	r := getRep()
	assert.NotNil(t, r.db)
	queryList := []string{"python", "Английский"}
	subdomain := ""
	for _, query := range queryList {
		books, err := r.GetByName(query, subdomain)
		assert.Nil(t, err)
		assert.NotEmptyf(t, len(books), "По '%s' запросу точно должны быть книги", query)
		for _, v := range books {
			assert.NotEmpty(t, v.Id, "Id книги не может быть пустым полем!")
			contains := strings.Contains(strings.ToLower(v.Name), strings.ToLower(query)) || strings.Contains(strings.ToLower(v.Description), strings.ToLower(query))
			assert.Truef(t, contains, "Название или описание книги должно содержать: '%s'", query)
		}
	}
	err := database.Close(r.db)
	assert.Nil(t, err)
}

func TestSaveBooks(t *testing.T) {
	r := getRep()
	assert.NotNil(t, r.db)

	data := models.QueryBooks{
		Query: "hello world",
		BookList: []models.Book{
			models.Book{
				Id: 1,
			},
		},
	}
	done, err := r.SaveBooks(data)
	assert.False(t, done)
	assert.EqualError(t, err, ErrorMsgNotExistBookId)

	var (
		query     = "Химические элементы"
		limit     = 3
		subdomain = "svoevagro"
	)

	books, err := r.GetByQuery(query, limit, subdomain)
	data = models.QueryBooks{
		Query:    "hello world",
		BookList: books,
	}
	done, err = r.SaveBooks(data)
	assert.True(t, done)
	assert.NoError(t, err)

	err = database.Close(r.db)
	assert.Nil(t, err)

}

func TestDeleteQuery(t *testing.T) {
	r := getRep()

	query := "hello world"
	done, err := r.DeleteQueryFromHistory(query)
	assert.True(t, done)
	assert.Nil(t, err)
	err = database.Close(r.db)
	assert.Nil(t, err)
}

func TestSaveUndefiendQuery(t *testing.T) {
	r := getRep()

	query := "sdaaisdianidnaindiasndnadnianidnaisndkansdknaskda"
	done, err := r.SaveUndefiendQuery(query)
	assert.True(t, done)
	assert.Nil(t, err)
	err = database.Close(r.db)
	assert.Nil(t, err)
}
