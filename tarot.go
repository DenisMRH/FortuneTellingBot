package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"
)

// Определяем структуру, которая будет представлять карту Таро
// Name - название карты, Description - её значение (около 300 символов)
type TarotCard struct {
	Name        string `json:"name"`        // Название карты, соответствует ключу "name" в JSON
	Description string `json:"description"` // Описание карты, соответствует ключу "description" в JSON
}

// Функция загрузки карт из JSON-файла
func loadTarotCards(filename string) ([]TarotCard, error) {
	// Читаем содержимое файла
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err // В случае ошибки возвращаем nil и ошибку
	}

	// Создаём слайс для хранения карт
	var cards []TarotCard

	// Разбираем JSON в слайс структур TarotCard
	err = json.Unmarshal(file, &cards)
	if err != nil {
		return nil, err // Если возникла ошибка при разборе, возвращаем nil и ошибку
	}

	return cards, nil // Возвращаем загруженные карты
}

// Функция выбора 3 случайных карт
func drawThreeCards(cards []TarotCard) []TarotCard {
	// Устанавливаем seed (инициализируем генератор случайных чисел)
	rand.Seed(time.Now().UnixNano()) // Используем текущее время в наносекундах, чтобы каждый раз был разный результат

	// Перемешиваем слайс карт случайным образом
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i] // Меняем местами элементы i и j
	})

	// Возвращаем первые три карты из перемешанного списка
	return cards[:3]
}
