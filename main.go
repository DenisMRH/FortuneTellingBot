package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// Структура запроса к DeepSeek API (Ollama)
type DeepSeekRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// Структура ответа от DeepSeek API (корректируем, если формат отличается)
type DeepSeekResponse struct {
	Response string `json:"response"`
}

// Функция для запроса к DeepSeek
func queryDeepSeek(prompt string) (string, error) {
	url := "http://localhost:11434/api/generate"

	// Формируем JSON-запрос
	reqBody, err := json.Marshal(DeepSeekRequest{
		Model:  "deepseek-r1:14b",
		Prompt: prompt,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка формирования запроса: %v", err)
	}

	// Отправка запроса
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ построчно (если JSONL)
	var fullResponse string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Ответ от DeepSeek (строка): %s", line)

		var dsResp DeepSeekResponse
		err = json.Unmarshal([]byte(line), &dsResp)
		if err != nil {
			log.Printf("Ошибка парсинга JSON: %v", err)
			continue
		}

		fullResponse += dsResp.Response
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return "", fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// Если ответ пустой, возвращаем сообщение по умолчанию
	if fullResponse == "" {
		return "DeepSeek не вернул ответ. Проверьте, работает ли модель.", nil
	}

	// Удаляем содержимое между <think> и </think> (с учётом переносов строк)
	cleanedResponse := regexp.MustCompile(`(?s)<think>.*?</think>`).ReplaceAllString(fullResponse, "")
	cleanedResponse = strings.TrimSpace(cleanedResponse) // Убираем лишние пробелы и пустые строки

	return cleanedResponse, nil
}

func main() {

	// Получаем API-ключ бота из переменной среды
	TELEGRAM_BOT_TOKEN := importEnv("hiddenFiles.env", "TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(TELEGRAM_BOT_TOKEN)
	if err != nil {
		// Если ошибка при создании бота - выводим её и завершаем работу
		log.Panic("Ошибка инициализации бота:", err)
	}
	bot.Debug = true
	log.Printf("Бот запущен как %s", bot.Self.UserName)

	// Настроим обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Обрабатываем входящие сообщения
	for update := range updates {
		if (update.Message == nil) || (len(update.Message.Text) > 170) { // Игнорируем не-текстовые сообщения
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы отправили слишком длинное сообщение, либо сообщение не текстовое.")
			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
			}
			continue
		}

		userPrompt := `
Ты профессиональная гадалка-таролог! Разбираешься во всех терминах тарологии, гороскопа и будущего!

Я напишу тебе вопрос касающийся моего будущего.  САМОЕ ГЛАВНОЕ ЕСЛИ МОЙ ВОПРОС НЕ БУДЕТ ПОХОЖ НА ЗАПРОС ПО ГАДАНИЮ, ТО ТЫ ДОЛЖЕН ВЕЖЛИВО ОТКАЗАТЬ В ГАДАНИИ И БОЛЬШЕ НИЧЕГО НЕ ПИСАТЬ!!! САМОЕ ГЛАВНОЕ ЕСЛИ МОЙ ВОПРОС НЕ БУДЕТ ПОХОЖ НА ЗАПРОС ПО ГАДАНИЮ, ТО ТЫ ДОЛЖЕН ВЕЖЛИВО ОТКАЗАТЬ В ГАДАНИИ И БОЛЬШЕ НИЧЕГО НЕ ПИСАТЬ!!!

Твоя задача:
Сделать мне расклад по трём картам таро. Карты ты должен выбрать сам, назвать и опиши что они значат. Рассказать как пройдёт мой день сегодня исходя из этих карт.
Расскажи не длинно, не больше 800 символов.
ОТВЕЧАЙ ТОЛЬКО НА РУССКОМ ЯЗЫКЕ НЕ ИСПОЛЬЗУЙ НИКАКИХ ДРУГИХ ЯЗЫКОВ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ!!!! ОТВЕЧАЙ ТОЛЬКО НА РУССКОМ ЯЗЫКЕ НЕ ИСПОЛЬЗУЙ НИКАКИХ ДРУГИХ ЯЗЫКОВ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ!!!!

Вопрос касающийся моего будущего, ОТВЕЧАЙ ТОЛЬКО НА РУССКОМ ЯЗЫКЕ НЕ ИСПОЛЬЗУЙ НИКАКИХ ДРУГИХ ЯЗЫКОВ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ!!!! САМОЕ ГЛАВНОЕ ЕСЛИ МОЙ ВОПРОС НЕ БУДЕТ ПОХОЖ НА ЗАПРОС ПО ГАДАНИЮ, ТО ТЫ ДОЛЖЕН ВЕЖЛИВО ОТКАЗАТЬ В ГАДАНИИ И БОЛЬШЕ НИЧЕГО НЕ ПИСАТЬ!!! :
` + update.Message.Text

		log.Printf("Сообщение от пользователя: %s", userPrompt)

		// Отправляем запрос в DeepSeek
		answer, err := queryDeepSeek(userPrompt)
		if err != nil {
			answer = "Ошибка при запросе к DeepSeek: " + err.Error()
		}

		// Проверяем, что ответ не пустой
		if answer == "" {
			answer = "Извините, я не смог обработать ваш запрос."
		}

		// Отправляем ответ пользователю
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, answer)
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
	}
}

func importEnv(fileName, varName string) (variable string) {
	err := godotenv.Load(fileName)
	if err != nil {
		log.Fatalf("Ошибка импорта файла %v", err)
	}

	variable = os.Getenv(varName)
	if variable == "" {
		log.Fatalf("Переменная %v не найдена.", variable)
	}
	fmt.Println("Переменная ", varName, " из файла ", fileName, " импортирована!")
	return
}
