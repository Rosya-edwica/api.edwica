package book

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
	stmt := `SELECT book.id as id, name, description, language, price, old_price, min_age, rating,
		year, image, url, currency, pages
		FROM book
		INNER JOIN query_to_book ON book.id = query_to_book.book_id
		WHERE query_to_book.query = $1`

	if limit > 0 {
		stmt += fmt.Sprintf(" LIMIT %d", limit)
	}

	err := r.db.Select(&rawBooks, stmt, query)
	if err != nil {
		logger.Log.Error("database.book.getByQuery:" + err.Error())
		return nil, errors.Wrap(err, "select book")
	}
	return models.NewBooks(rawBooks), nil
}

// Используем этот метод, если по запросу не нашлось записей в таблице query_to_book
func (r *Repository) GetByName(query, subdomain string) ([]models.Book, error) {
	fmt.Println("get by nam")
	rawBooks := make([]entities.Book, 0)
	var dbQuery string
	dbQuery = `SELECT id, name, description, language, price, old_price, min_age, rating,
		year, image, url, currency, pages
		FROM book
		WHERE name ILIKE $1 AND source = $2`
	err := r.db.Select(&rawBooks, dbQuery, "%"+query+"%", subdomain)
	fmt.Println(len(rawBooks), query)
	if err != nil {
		logger.Log.Error("database.book.getByName:" + err.Error())
		return []models.Book{}, errors.Wrap(err, "select book by name")
	}
	return models.NewBooks(rawBooks), nil
}

func (r *Repository) SaveNewBooks(data models.QueryBooks) (done bool, err error) {
	if len(data.BookList) == 0 {
		return false, nil
	}
	valuesQuery := make([]string, 0, len(data.BookList))
	valuesArgs := make([]interface{}, 0, len(data.BookList))
	id := 0
	for _, b := range data.BookList {
		valuesQuery = append(valuesQuery, fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, 
			$%d, $%d, $%d, $%d)`, id+1, id+2, id+3, id+4, id+5, id+6, id+7, id+8, id+9, id+10, id+11, id+12, id+13))
		valuesArgs = append(valuesArgs, b.Id, b.Name, b.Description, b.Image, b.Url, b.OldPrice, b.Price, b.Currency,
			b.MinAge, b.Language, b.Rating, b.Pages, b.Year)
		id += 13
		fmt.Println(b.Id, b.Name, b.Price)
	}
	query := fmt.Sprintf(`INSERT INTO book(
                 id, name, description, image, url, old_price, price, currency, min_age, 
                 language, rating, pages, year) VALUES %s ON CONFLICT DO NOTHING`, strings.Join(valuesQuery, ","))
	_, err = r.db.Exec(query, valuesArgs...)
	if err != nil {
		return false, errors.Wrap(err, "ошибка при сохранении новых книг")
	}
	return r.SaveBooksToQuery(data)
}

func (r *Repository) SaveBooksToQuery(data models.QueryBooks) (done bool, err error) {
	var bookIds []string
	for _, i := range data.BookList {
		bookIds = append(bookIds, fmt.Sprintf("('%s', %d)", strings.ToLower(data.Query), i.Id))
	}
	//Связываем книги с запросом в таблице истории
	_, err = r.db.Exec(fmt.Sprintf("INSERT INTO query_to_book(query, book_id) VALUES %s", strings.Join(bookIds, ",")))
	if err != nil {
		return false, err
	}

	return true, nil
}

// Сохраняем запросы, для которых не нашлось книг. Чтобы в будущем не тратить на них время
func (r *Repository) SaveUndefiendQuery(query string) (bool, error) {
	_, err := r.db.Exec(fmt.Sprintf("INSERT INTO query_to_book(query, book_id) VALUES('%s', NULL) ON CONFLICT DO NOTHING", query))
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
