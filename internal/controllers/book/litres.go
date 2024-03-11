package book

import (
	"encoding/json"
	"fmt"
	"github.com/Rosya-edwica/api.edwica/internal/entities"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/go-faster/errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

func SearchBooks(query string, limit int) (result models.QueryBooks, err error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("limit", strconv.Itoa(limit))
	params.Add("types", "text_book")

	queryUrl := "https://api.litres.ru/foundation/api/search?" + params.Encode()
	resp, err := decodeJsonResponse(queryUrl, nil, &entities.LitresSearchResponse{}, "GET")
	if resp == nil || err != nil {
		return models.QueryBooks{}, errors.Wrap(err, fmt.Sprintf("Ничего не нашлось по запросу: '%s'", query))
	}
	data := resp.(*entities.LitresSearchResponse)
	books := make([]entities.LitresBook, 0, len(data.Response.Books))
	wg := sync.WaitGroup{}
	wg.Add(cap(books))
	for _, i := range data.Response.Books {
		go func(id int) {
			defer wg.Done()
			book, err := parseBook(id)
			fmt.Println(book.Name, err)
			books = append(books, book)
		}(i.Book.Id)
	}
	wg.Wait()
	return models.QueryBooks{
		Query:    query,
		BookList: models.ConvertLitresBooks(books),
	}, nil
}

func parseBook(bookId int) (entities.LitresBook, error) {
	queryUrl := fmt.Sprintf("https://api.litres.ru/foundation/api/arts/%d", bookId)
	fmt.Println(queryUrl)
	resp, err := decodeJsonResponse(queryUrl, nil, &entities.LitresBookResponse{}, "GET")
	if resp == nil || err != nil {
		return entities.LitresBook{}, errors.Wrap(err, fmt.Sprintf("Ничего не нашлось по запросу: '%d'", bookId))
	}
	data := resp.(*entities.LitresBookResponse)
	book := data.Response.Books.LitresBook
	return book, nil
}

// decodeJsonResponse старается распарсить любой JSON в структуру dataStruct.
// dataStruct - &структуры, в которую нужно распарсить JSON.
// Возвращает заполненную структуру и ошибку
func decodeJsonResponse(url string, headers map[string]string, dataStruct interface{}, method string) (data interface{}, err error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Ошибка при создании запроса: "+url)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := client.Do(request)
	if err != nil || response.StatusCode != 200 {
		return nil, errors.Wrap(err, "Ошибка при отправке запроса: "+url)
	}

	err = json.NewDecoder(response.Body).Decode(dataStruct)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("decodeJsonResponse by url:%s err:%s", url, err))
		return nil, errors.Wrap(err, "Ошибка при наполнении структуры")
	}
	err = response.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "Не удалось закрыть тело ответа")
	}
	return dataStruct, nil
}
