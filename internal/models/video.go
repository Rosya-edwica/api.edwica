package models

import "github.com/Rosya-edwica/api.edwica/internal/entities"

type Video struct {
	Id    string `json:"id"` // gorm:"primaryKey"
	Name  string `json:"name"`
	Url   string `json:"link"`
	Image string `json:"header_image"`
}

type QueryVideos struct {
	Query     string  `json:"skill"`
	VideoList []Video `json:"materials"`
}

func NewVideos(rawVideos []entities.Video) (videos []Video) {
	for _, item := range rawVideos {
		videos = append(videos, Video{
			Id:    item.Id,
			Name:  item.Name,
			Url:   item.Url,
			Image: item.Image,
		})
	}
	return videos
}
