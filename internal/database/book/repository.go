package book

import (
	"fmt"
	"log"
	"strings"

	"github.com/Rosya-edwica/api.edwica/internal/entities"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/go-sql-driver/mysql"

	"github.com/go-faster/errors"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

const (
	ErrorMsgNotExistBookId = "book is does not exist"
)

func NewRepository(db *sqlx.DB) *Repository {
	if db == nil {
		log.Fatal("Передали в репозиторий неправильное подключение")
	}
	return &Repository{
		db: db,
	}
}

// Сначала ищем книги в query_to_book - это наш кэш
func (r *Repository) GetByQuery(query string, limit int, subdomain string) ([]models.Book, error) {
	rawBooks := make([]entities.Book, 0)
	var dbQuery string
	if subdomain != "" {
		dbQuery = fmt.Sprintf(`SELECT book.id as id, name, description, language, price, old_price, min_age, rating,
		year, image, url, currency, pages, is_audio
		FROM book
		INNER JOIN query_to_book ON book.id = query_to_book.book_id
		WHERE query_to_book.query = ? AND subdomain = '%s'
		`, subdomain)
	} else {
		dbQuery = `SELECT book.id as id, name, description, language, price, old_price, min_age, rating,
		year, image, url, currency, pages, is_audio
		FROM book
		INNER JOIN query_to_book ON book.id = query_to_book.book_id
		WHERE query_to_book.query = ? AND price = 0`
	}

	if limit > 0 {
		dbQuery = fmt.Sprintf("%s LIMIT %d", dbQuery, limit)
	}
	err := r.db.Select(&rawBooks, dbQuery, query)
	if err != nil {
		logger.Log.Error("database.book.getByQuery:" + err.Error())
		return nil, errors.Wrap(err, "select book")
	}
	return models.NewBooks(rawBooks), nil
}

// Используем этот метод, если по запросу не нашлось записей в таблице query_to_book
func (r *Repository) GetByName(query, subdomain string) ([]models.Book, error) {
	rawBooks := make([]entities.Book, 0)
	var dbQuery string
	if subdomain != "" {
		dbQuery = fmt.Sprintf(`SELECT id, name, description, language, price, old_price, min_age, rating,
		year, image, url, currency, pages, is_audio
		FROM book
		WHERE MATCH (name, description) AGAINST (?) AND subdomain = '%s'`, subdomain)
	} else {
		dbQuery = `SELECT id, name, description, language, price, old_price, min_age, rating,
		year, image, url, currency, pages, is_audio
		FROM book
		WHERE MATCH (name, description) AGAINST (?) AND price = 0`
	}

	err := r.db.Select(&rawBooks, dbQuery, query)
	if err != nil {
		logger.Log.Error("database.book.getByName:" + err.Error())
		return nil, errors.Wrap(err, "select book by name")
	}
	return models.NewBooks(rawBooks), nil
}

func (r *Repository) GetByNameWithSubdomain(query string, subdomain string) ([]models.Book, error) {
	rawBooks := make([]entities.Book, 0)
	dbQuery := `SELECT id, name, description, language, price, old_price, min_age, rating,
		year, image, url, currency, pages, is_audio
		FROM book
		WHERE MATCH (name, description) AGAINST (?) AND subdomain = ?`

	err := r.db.Select(&rawBooks, dbQuery, query, subdomain)
	if err != nil {
		logger.Log.Error("database.book.getByName:" + err.Error())
		return nil, errors.Wrap(err, "select book by name")
	}
	fmt.Println(rawBooks)
	return models.NewBooks(rawBooks), nil
}

// Убрать транзакцию на простой инсерт
// Транзакция сохранения новых видосов и связи запроса с видосами
func (r *Repository) SaveBooks(data models.QueryBooks) (done bool, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		logger.Log.Error("database.book.saveBooks- start transaction:" + err.Error())
		return false, errors.Wrap(err, "book saving failed transaction")
	}
	defer func() {
		if err != nil {
			errRb := tx.Rollback()
			if errRb != nil {
				logger.Log.Error("database.book.saveBooks- failed stop transaction:" + err.Error())
				err = errors.Wrap(err, "book saving during rollback")
				return
			}

			return
		}
		err = tx.Commit()
		if err != nil {
			logger.Log.Error("database.book.saveBooks- commit transaction:" + err.Error())
		}
	}()

	var bookIds []string
	for _, i := range data.BookList {
		bookIds = append(bookIds, fmt.Sprintf("('%s', %d)", strings.ToLower(data.Query), i.Id))
	}
	//Связываем книги с запросом в таблице истории
	_, err = tx.Exec(fmt.Sprintf("INSERT INTO query_to_book(query, book_id) VALUES %s", strings.Join(bookIds, ",")))
	if err != nil {
		msErr, _ := err.(*mysql.MySQLError)
		if msErr.Number == 1452 {
			return false, errors.New(ErrorMsgNotExistBookId)
		}
		return false, err
	}

	return true, nil
}

// Сохраняем запросы, для которых не нашлось книг. Чтобы в будущем не тратить на них время
func (r *Repository) SaveUndefiendQuery(query string) (bool, error) {
	_, err := r.db.Exec(fmt.Sprintf("INSERT INTO query_to_book(query, book_id) VALUES('%s', NULL)", query))
	if err != nil {
		logger.Log.Error("database.book.SaveUndefiendQuery:" + err.Error())
		return false, err
	}

	return true, nil
}

// Удаляем ненужные связи - по сути этот метод создан только для тестов
func (r *Repository) DeleteQueryFromHistory(query string) (bool, error) {
	_, err := r.db.Exec(fmt.Sprintf("DELETE FROM query_to_book WHERE query = '%s'", query))
	if err != nil {
		logger.Log.Error("database.book.DeleteQueryFromHistory:" + err.Error())
		return false, err
	}
	return true, nil
}
