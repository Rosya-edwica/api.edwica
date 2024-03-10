package entities

type Book struct {
	Id          int     `db:"id"`
	Name        string  `db:"name"`
	Description string  `db:"description"`
	Image       string  `db:"image"`
	Url         string  `db:"url"`
	OldPrice    float32 `db:"old_price"`
	Price       float32 `db:"price"`
	Currency    string  `db:"currency"`
	MinAge      int     `db:"min_age"`
	Language    string  `db:"language"`
	Rating      float32 `db:"rating"`
	Pages       int     `db:"pages"`
	Year        int     `db:"year"`
}

type LitresSearchResponse struct {
	Response struct {
		Books []struct {
			Book struct {
				Id int `json:"id"`
			} `json:"instance"`
		} `json:"data"`
	} `json:"payload"`
}

type LitresBookResponse struct {
	Response struct {
		Books struct {
			LitresBook
		} `json:"data"`
	} `json:"payload"`
}

type LitresBook struct {
	Id          int    `json:"id"`
	Name        string `json:"title"`
	Description string `json:"html_annotation"`
	Image       string `json:"cover_url"`
	Url         string `json:"url"`
	Price       struct {
		Full     float32 `json:"full_price"`
		Final    float32 `json:"final_price"`
		Currency string  `json:"currency"`
	} `json:"prices"`
	MinAge   int    `json:"min_age"`
	Language string `json:"language_code"`
	Rating   struct {
		Rate float32 `json:"rated_avg"`
	} `json:"rating"`
	Additional struct {
		Pages int `json:"current_pages_or_seconds"`
	} `json:"additional_info"`
	Date string `json:"date_written_at"`
}
