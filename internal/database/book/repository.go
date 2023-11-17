package book

import (
	"fmt"
	"github.com/Rosya-edwica/api.edwica/internal/entities"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"log"
	"strings"

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

// Сначала ищем книги в query_to_book - это наш кэш
func (r *Repository) GetByQuery(query string, limit int) ([]models.Book, error) {
	rawBooks := make([]entities.Book, 0)
	dbQuery := `SELECT book.id as id, name, description, language, price, old_price, min_age, rating,
	year, image, url, currency, pages, is_audio
	FROM book
	INNER JOIN query_to_book ON book.id = query_to_book.book_id
	WHERE LOWER(query_to_book.query) = ?
	`
	if limit > 0 {
		dbQuery = fmt.Sprintf("%s LIMIT %d", dbQuery, limit)
	}
	err := r.db.Select(&rawBooks, dbQuery, query)
	if err != nil {
		return nil, errors.Wrap(err, "select book")
	}
	return models.NewBooks(rawBooks), nil
}

// Используем этот метод, если по запросу не нашлось записей в таблице query_to_book
func (r *Repository) GetByName(query string, limit int) ([]models.Book, error) {
	rawBooks := make([]entities.Book, 0)
	dbQuery := `SELECT id, name, description, language, price, old_price, min_age, rating,
		year, image, url, currency, pages, is_audio
		FROM book 
		WHERE MATCH (name, description) AGAINST (?)`
	if limit > 0 {
		dbQuery = fmt.Sprintf("%s LIMIT %d", dbQuery, limit)
	}
	err := r.db.Select(&rawBooks, dbQuery, query)
	if err != nil {
		return nil, errors.Wrap(err, "select book by name")
	}
	return models.NewBooks(rawBooks), nil
}

// Убрать транзакцию на простой инсерт
// Транзакция сохранения новых видосов и связи запроса с видосами
func (r *Repository) SaveBooks(data models.QueryBooks) (done bool, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return false, errors.Wrap(err, "book saving failed transaction")
	}
	defer func() {
		if err != nil {
			errRb := tx.Rollback()
			if errRb != nil {
				err = errors.Wrap(err, "book saving during rollback")
				return
			}

			return
		}
		err = tx.Commit()
	}()

	var bookIds []string
	for _, i := range data.BookList {
		bookIds = append(bookIds, fmt.Sprintf("('%s', %d)", data.Query, i.Id))
	}
	//Связываем книги с запросом в таблице истории
	_, err = tx.Exec(fmt.Sprintf("INSERT IGNORE INTO query_to_book(query, book_id) VALUES %s", strings.Join(bookIds, ",")))
	fmt.Println(err, "fsdfdsfdsfsdfsdfdsf")
	if err != nil {
		return false, err
	}

	return true, nil
}
