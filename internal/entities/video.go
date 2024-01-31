package entities

type Video struct {
	Id    string `db:"id"` // gorm:"primaryKey"
	Name  string `db:"name"`
	Url   string `db:"url"`
	Image string `db:"img"`
}

type QueryVideo struct {
}
