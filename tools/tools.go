package tools

import (
	"net/url"
)

// Выбираем только уникальные параметры в запросе и ведем поиск по ним
// Например: Для api/v1/videos?text=golang&text=python&text=golang будет создано две горутины - одна для python, другая для golang
// Если не использовать этот метод, то будет три горутины, две из которых будут одинаковыми
func UniqueSlice(input []string) []string {
	var output []string
	allKeys := make(map[string]bool)
	for _, item := range input {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			output = append(output, item)
		}
	}
	return output
}

// Функция, чтобы мы могли нормально прочитать "C++" вместо "С  "
// Не работает для других символов C# *
func DecodeAllQueries(items []string) []string {
	var decoded []string
	for _, s := range items {
		text := url.QueryEscape(s)
		decoded = append(decoded, text)
	}
	return decoded
}
