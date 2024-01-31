package models

import (
	"github.com/Rosya-edwica/api.edwica/internal/entities"
)

type Book struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"header_image"`
	Url         string  `json:"link"`
	IsAudio     bool    `json:"is_audio"`
	OldPrice    int     `json:"old_price"`
	Price       int     `json:"price"`
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

func NewBooks(rawBooks []entities.Book) (books []Book) {
	for _, i := range rawBooks {
		books = append(books, Book{
			Id:          i.Id,
			Name:        i.Name,
			Description: i.Description,
			Image:       i.Image,
			Url:         i.Url,
			IsAudio:     i.IsAudio,
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
