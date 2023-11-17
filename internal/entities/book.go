package entities

type Book struct {
	Id          int     `db:"id"`
	Name        string  `db:"name"`
	Description string  `db:"description"`
	Image       string  `db:"image"`
	Url         string  `db:"url"`
	IsAudio     bool    `db:"is_audio"`
	OldPrice    int     `db:"old_price"`
	Price       int     `db:"price"`
	Currency    string  `db:"currency"`
	MinAge      int     `db:"min_age"`
	Language    string  `db:"language"`
	Rating      float32 `db:"rating"`
	Pages       int     `db:"pages"`
	Year        int     `db:"year"`
}
