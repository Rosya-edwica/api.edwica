// Здесь мы подбираем новые видео для запросов, у которых нет видео в БД. Для этого:
/*
1. Создаем ссылку, которую бы нам выдал youtube, если бы мы с компа зашли на сайт и вбили туда наш запрос
2. Ютуб хранит JSON-ответ в HTML-коде в блоках <script>. Нужный нам JSON лежит в 5 блоке по индексу 4
3. Нам нужно распарсить блок <script>. Нужно вырезать весь код и оставить только основной блок JSON с информацией о видео. Инфу о плейлистах, каналах и тд вырезаем
4. JSON нужно подогнать под нашу структуру Youtube
5. После чего распарсить структуру и вернуть ответ в виде списка структуры Video
*/
package video

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/go-faster/errors"
)

func GetUndiscoveredVideosByAPI(queryList []string, limit int) (response []models.QueryVideos, errors []error) {
	wg := sync.WaitGroup{}
	wg.Add(len(queryList))

	for _, i := range queryList {
		go func(query string) {
			defer wg.Done()
			data, err := getVideoByAPI(query, limit)
			if err != nil {
				errors = append(errors, err)
				return
			}
			response = append(response, data)
		}(i)
	}
	wg.Wait()
	return
}

func getVideoByAPI(query string, limit int) (response models.QueryVideos, err error) {
	baseURL := "https://www.googleapis.com/youtube/v3/search"
	baseVideoURL := "https://www.youtube.com/watch?v"
	params := url.Values{}
	params.Add("part", "snippet")
	params.Add("q", "python")                                    // замените на ваш поисковый запрос
	params.Add("key", "AIzaSyCB_uQD2jnxPzYqSR92CtSpTwpHhUYv0Uc") // замените на ваш API ключ
	params.Add("maxResults", "10")
	params.Add("regionCode", "RU")
	params.Add("type", "video")
	resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return models.QueryVideos{}, errors.Wrap(err, "select video by api")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return models.QueryVideos{}, errors.Wrap(err, "read body youtube api")
	}
	var (
		data   Response
		videos []models.Video
	)
	json.Unmarshal(body, &data)

	for _, i := range data.Videos {
		videos = append(videos, models.Video{
			Id:    i.Id.VideoId,
			Name:  i.Snippet.Title,
			Image: i.Snippet.Image.Default.Url,
			Url:   fmt.Sprintf("%s=%s", baseVideoURL, i.Id.VideoId),
		})
	}
	response = models.QueryVideos{
		Query:     query,
		VideoList: videos,
	}
	VideoCacheMutex.Lock()
	VideoCache[query] = videos
	VideoCacheMutex.Unlock()
	return response, nil
}

func GetUndiscoveredVideos(queryList []string, limit int) (response []models.QueryVideos, errors []error) {
	wg := sync.WaitGroup{}
	wg.Add(len(queryList))
	for _, query := range queryList {
		go func(query string) {
			defer wg.Done()
			videos, err := GetVideosFromYoutube(query)
			if len(videos) >= limit {
				videos = videos[:limit]
			}
			if err != nil {
				errors = append(errors, err)
			} else {
				response = append(response, models.QueryVideos{
					Query:     query,
					VideoList: videos,
				})
				VideoCacheMutex.Lock()
				VideoCache[query] = videos
				VideoCacheMutex.Unlock()

			}
		}(query)
	}
	wg.Wait()
	return
}

func GetVideosFromYoutube(query string) ([]models.Video, error) {
	query = strings.Join(strings.Split(query, " "), "+")                                  // Если передаем навык из двух слов, то нужно заменить пробел на +
	encodedUrl := "https://www.youtube.com/results?search_query=" + url.PathEscape(query) // Кодируем кириллицу, чтобы не было кракозябр
	html, err := GetHTML(encodedUrl)
	if err != nil {
		return nil, err
	}
	youtubeData, err := ReadHTMLContent(html)
	if err != nil {
		return nil, err
	}
	videos := convertYoutubeJsonToVideoList(&youtubeData)
	return videos, nil

}

func GetHTML(link string) (string, error) {
	req, err := http.Get(link)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()
	content, err := io.ReadAll(req.Body)
	if err != nil {
		return "", errors.Wrap(err, "reading HTML content:"+link)
	}
	return string(content), nil
}

// Мы находим в верстке Ютуба скрипт с JSON-данными поиска. Вырезаем все лишнее с помощью регулярок, а нужный JSON кладем в структуру
func ReadHTMLContent(content string) (Youtube, error) {
	var (
		reScripts             = regexp.MustCompile(`\<script.*\<\/script\>`)                                             // Выбираем только блоки script из всего HTML
		reJsonData            = regexp.MustCompile(`{"itemSectionRenderer":{"contents":.*,{"continuationItemRenderer":`) // Вырезаем только самое необходимое в JSON
		reCurrentData         = regexp.MustCompile(`.*}}]`)                                                              // Отрезаем лишнее
		currentScriptPosition = 4
		currentScript         string
		youtubeData           Youtube
	)
	scripts := reScripts.FindAllString(content, -1)
	if len(scripts) >= currentScriptPosition {
		currentScript = scripts[currentScriptPosition]
	} else {
		return Youtube{}, errors.New("В HTML-коде нет нужного блока с JSON")
	}

	jsonData := reJsonData.FindString(currentScript)
	currentJson := reCurrentData.FindString(jsonData) + "}}" // Добавляем }} чтобы получить валидный JSON, который мы обрезали
	err := json.Unmarshal([]byte(currentJson), &youtubeData)
	if err != nil {
		return Youtube{}, errors.Wrap(err, "converting json to struct")
	}
	return youtubeData, nil
}

func convertYoutubeJsonToVideoList(y *Youtube) []models.Video {
	var videos []models.Video
	for _, item := range y.Data.Contents {
		if item.Video.Id == "" {
			continue
		}
		video := models.Video{
			Id:    item.Video.Id,
			Name:  item.Video.Title.Runs[0].Text,
			Image: item.Video.Image.Items[len(item.Video.Image.Items)-1].Url,
			Url:   "https://www.youtube.com/watch?v=" + item.Video.Id,
		}
		videos = append(videos, video)
	}
	return videos
}
