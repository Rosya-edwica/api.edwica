package video

import (
	"fmt"
	"log"
	"strings"

	"github.com/Rosya-edwica/api.edwica/internal/entities"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"

	"github.com/go-faster/errors"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	if db == nil {
		log.Fatal("Передали в репозиторий неправильное подключение")
	}
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetByQuery(query string, limit int) ([]models.Video, error) {
	rawVideos := make([]entities.Video, 0)
	// Намерено делаю поле img пустым, так как оно не используется в АПИ,
	// но в таблице влияет на уникальность данных (у некоторых видео разные картинки по размеру)
	dbQuery := fmt.Sprintf(`SELECT DISTINCT video.video_id as id, video.name as name, video.url as url, '' as img
	FROM video
	INNER JOIN query_to_video ON video.video_id = query_to_video.video_id
	WHERE query_to_video.query = '%s'`, query)
	if limit > 0 {
		dbQuery = fmt.Sprintf("%s LIMIT %d", dbQuery, limit)
	}
	err := r.db.Select(&rawVideos, dbQuery)
	if err != nil {
		logger.Log.Error("database.video.getByQuery:" + err.Error())
		return nil, errors.Wrap(err, "select video")
	}
	return models.NewVideos(rawVideos), nil
}

// Транзакция сохранения новых видосов и связи запроса с видосами
func (r *Repository) SaveVideos(data models.QueryVideos) (done bool, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		logger.Log.Error("database.video.saveVideos - start transaction:" + err.Error())
		return false, errors.Wrap(err, "video saving failed transaction")
	}
	defer func() {
		if err != nil {
			errRb := tx.Rollback()
			if errRb != nil {
				err = errors.Wrap(err, "video saving during rollback")
				logger.Log.Error("database.video.saveVideos - stop transaction:" + err.Error())
				return
			}
			return
		}
		err = tx.Commit()
		if err != nil {
			logger.Log.Error("database.video.saveVideos - commit transaction:" + err.Error())

		}
	}()
	err = r.saveVideos(tx, data.VideoList)
	if err != nil {
		return false, err // Если не удалось сохранить видосы, значит запрос не с чем связывать
	}
	videoIds := []string{}
	for _, i := range data.VideoList {
		if i.Id == "" {
			continue
		}
		videoIds = append(videoIds, fmt.Sprintf("('%s', '%s')", strings.ToLower(data.Query), i.Id))
	}
	if len(videoIds) == 0 {
		return false, err
	}
	//Связываем видосы с запросом в таблице истории
	_, err = tx.Exec(fmt.Sprintf("INSERT INTO query_to_video(query, video_id) VALUES %s ON CONFLICT DO NOTHING", strings.Join(videoIds, ",")))
	if err != nil {
		logger.Log.Error("database.video.saveVideos - insert data to query_to_video:" + err.Error())
		return false, err
	}

	return true, nil
}

// Сохраняем видосы
func (r *Repository) saveVideos(tx *sqlx.Tx, videos []models.Video) error {
	if len(videos) == 0 {
		// Вообще надо сохранить как не определенные
		return nil
	}
	valuesQuery := make([]string, 0, len(videos))
	valuesArgs := make([]interface{}, 0, len(videos))

	var id int
	for _, video := range videos {
		valuesQuery = append(valuesQuery, fmt.Sprintf("($%d, $%d, $%d, $%d)", id+1, id+2, id+3, id+4))
		valuesArgs = append(valuesArgs, video.Id, video.Name, video.Url, video.Image)
		id += 4
	}
	query := fmt.Sprintf(`INSERT INTO video(video_id, name, url, img)	VALUES %s ON CONFLICT DO NOTHING`, strings.Join(valuesQuery, ","))
	_, err := tx.Exec(query, valuesArgs...)
	if err != nil {
		logger.Log.Error("database.video.saveVideos - insert new videos:" + err.Error())
		return errors.Wrap(err, "adding videos to db")
	}
	return nil
}

func (r *Repository) DeleteVideoById(id string) (bool, error) {
	_, err := r.db.Exec(fmt.Sprintf("DELETE FROM video WHERE id = '%s'", id))
	if err != nil {
		logger.Log.Error("database.video.deleteVideos:" + err.Error())
		return false, err
	}
	return true, nil
}
