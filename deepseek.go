package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Структура запроса к DeepSeek API (Ollama)
type DeepSeekRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// Структура ответа от DeepSeek API
type DeepSeekResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// Функция запроса к DeepSeek API
func queryDeepSeek(prompt string) (string, error) {
	url := "http://localhost:11434/v1/completions"

	// Формируем JSON-запрос
	reqBody, err := json.Marshal(map[string]interface{}{
		"model":  "deepseek-r1:1.5b",
		"prompt": prompt,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка формирования запроса: %v", err)
	}

	// Отправляем HTTP-запрос
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	// Декодируем JSON-ответ
	var dsResp DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil {
		return "", fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	// Проверяем, есть ли текст в `choices`
	if len(dsResp.Choices) == 0 {
		return "DeepSeek не вернул текстовый ответ.", nil
	}

	// Убираем <think> ... </think> (если есть)
	responseText := regexp.MustCompile(`(?s)<think>.*?</think>`).ReplaceAllString(dsResp.Choices[0].Text, "")
	responseText = strings.TrimSpace(responseText)

	// Возвращаем обработанный ответ
	return responseText, nil
}
