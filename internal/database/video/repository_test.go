package video

import (
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

func TestGetByQuery(t *testing.T) {
	r := getRep()

	query := "golang"
	limit := 5

	videos, err := r.GetByQuery(query, limit)
	assert.Nil(t, err)
	assert.NotEmptyf(t, videos, "По '%s' запросу точно должны быть видосы", query)
	for _, v := range videos {
		assert.NotEmpty(t, v.Id, "Id видео не может быть пустым полем!")
		// contains := strings.Contains(, strings.ToLower(query)) || strings.Contains(strings.ToLower(v.Description), strings.ToLower(query))
	}
	err = database.Close(r.db)
	assert.Nil(t, err)
}

func TestSaveVideos(t *testing.T) {
	r := getRep()
	data := models.QueryVideos{
		Query: "hello world",
		VideoList: []models.Video{
			models.Video{
				Id:    "aaaa",
				Name:  "bbb",
				Image: "ccc",
				Url:   "ddd",
			},
		},
	}
	done, err := r.SaveVideos(data)
	assert.Nil(t, err)
	assert.True(t, done)

	for _, v := range data.VideoList {
		done, err = r.DeleteVideoById(v.Id)
		assert.Nil(t, err)
		assert.True(t, done)
	}

	videos, err := r.GetByQuery("hello world", 3)
	assert.Nil(t, err)
	assert.Empty(t, videos)

	err = database.Close(r.db)
	assert.Nil(t, err)
}
