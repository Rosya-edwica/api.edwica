package models

import (
	"github.com/Rosya-edwica/api.edwica/internal/entities"
	"strconv"
	"strings"
)

type Book struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"header_image"`
	Url         string  `json:"link"`
	OldPrice    float32 `json:"old_price"`
	Price       float32 `json:"price"`
	Currency    string  `json:"currency"`
	MinAge      int     `json:"min_age"`
	Language    string  `json:"language"`
	Rating      float32 `json:"rating"`
	Pages       int     `json:"pages"`
	Year        int     `json:"year"`
}

type QueryBooks struct {
	Query    string `json:"skill"`
	BookList []Book `json:"materials"`
}

type BookParams struct {
	Texts  []string
	Source string
	Limit  int
}

func NewBooks(rawBooks []entities.Book) []Book {
	books := make([]Book, 0, len(rawBooks))
	for _, i := range rawBooks {
		books = append(books, Book{
			Id:          i.Id,
			Name:        i.Name,
			Description: i.Description,
			Image:       i.Image,
			Url:         i.Url,
			OldPrice:    i.OldPrice,
			Price:       i.Price,
			Currency:    i.Currency,
			MinAge:      i.MinAge,
			Language:    i.Language,
			Rating:      i.Rating,
			Pages:       i.Pages,
			Year:        i.Year,
		})
	}
	return books
}

func ConvertLitresBooks(rawBooks []entities.LitresBook) (books []Book) {
	for _, i := range rawBooks {
		var year int
		if i.Date != "" {
			year, _ = strconv.Atoi(strings.Split(i.Date, "-")[0])
		}
		books = append(books, Book{
			Id:          i.Id,
			Name:        i.Name,
			Description: i.Description,
			Image:       i.Image,
			Url:         i.Url,
			OldPrice:    i.Price.Full,
			Price:       i.Price.Final,
			Currency:    i.Price.Currency,
			MinAge:      i.MinAge,
			Language:    i.Language,
			Rating:      i.Rating.Rate,
			Pages:       i.Additional.Pages,
			Year:        year,
		})
	}
	return
}
