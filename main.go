package main // Основной пакет приложения

import (
	// Стандартная библиотека для логирования
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	log "log"
	"net/http"
	"os"
	"regexp"
	"strings"

	// Библиотека для работы с Telegram Bot API
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// Глобальная переменная для хранения ID последнего сообщения, отправленного ботом.
// Используется для удаления предыдущего сообщения бота перед отправкой нового.
var lastBotMessageID int

// Глобальная мапа для хранения состояния каждого чата (ключ — ID чата, значение — строка состояния).
// Возможные состояния: "main", "question", "instruction", "tariffs".
// "main" — главное меню; "question" — режим для ввода вопроса; "instruction"/"tariffs" — режимы просмотра инструкций и тарифов.
var userState = make(map[int64]string)

// Функция main — точка входа в программу.
func main() {
	// Получаем API-ключ бота из переменной среды
	TELEGRAM_BOT_TOKEN := importEnv("hiddenFiles.env", "TELEGRAM_BOT_TOKEN")

	// Создаем нового бота, используя ваш уникальный токен.
	bot, err := tgbotapi.NewBotAPI(TELEGRAM_BOT_TOKEN)
	// Если произошла ошибка (например, неверный токен), логируем ошибку и завершаем выполнение.
	if err != nil {
		log.Panic(err)
	}

	// Включаем режим отладки для подробного логирования работы бота.
	// bot.Debug = true

	// Выводим в лог имя авторизованного аккаунта бота.
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Создаем объект конфигурации для получения обновлений.
	// Значение 0 означает, что мы хотим получать все обновления с самого начала.
	updateConfig := tgbotapi.NewUpdate(0)
	// Устанавливаем таймаут в 60 секунд для ожидания новых обновлений.
	updateConfig.Timeout = 60

	// Получаем канал, по которому будут поступать обновления (новые сообщения).
	updates := bot.GetUpdatesChan(updateConfig)

	// Бесконечный цикл для обработки каждого обновления.
	for update := range updates {
		// Если обновление содержит сообщение (а не, например, callback-запрос), то:
		if update.Message != nil {
			// Передаем сообщение в функцию handleMessage для обработки.
			handleMessage(bot, update.Message)
		}
	}
}

// Функция handleMessage обрабатывает входящие сообщения от пользователей.
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// ----------------------- Удаление сообщения пользователя -----------------------
	// Создаем объект для удаления сообщения пользователя по ID чата и ID сообщения.
	deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
	// Отправляем запрос на удаление сообщения пользователя.
	bot.Send(deleteMsg)

	// ----------------------- Удаление предыдущего сообщения бота -----------------------
	// Если переменная lastBotMessageID не равна 0, значит бот уже отправлял сообщение.
	if lastBotMessageID != 0 {
		// Создаем запрос на удаление предыдущего сообщения бота.
		deleteBotMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, lastBotMessageID)
		// Отправляем запрос на удаление.
		bot.Send(deleteBotMsg)
	}

	// ----------------------- Определение состояния пользователя -----------------------
	// Извлекаем состояние текущего чата из мапы userState.
	state, exists := userState[message.Chat.ID]
	// Если состояние не задано, по умолчанию считаем, что чат в главном меню.
	if !exists {
		state = "main"
		userState[message.Chat.ID] = "main"
	}

	// ----------------------- Обработка сообщения в зависимости от состояния -----------------------
	switch state {
	// Состояние "main" — пользователь находится в главном меню.
	case "main":
		switch message.Text {
		// Команда /start инициирует главное меню.
		case "/start":
			// Состояние остается "main" и отправляется главное меню.
			lastBotMessageID = sendMainMenu(bot, message.Chat.ID)
		// При выборе пункта "Задать вопрос" переходим в режим вопроса.
		case "Задать вопрос":
			// Устанавливаем состояние для данного чата в "question".
			userState[message.Chat.ID] = "question"
			// Отправляем меню для вопросов, где пользователь может выбрать вариант или ввести вопрос вручную.
			lastBotMessageID = sendQuestionMenu(bot, message.Chat.ID)
		// При выборе "Инструкция" переключаем состояние на "instruction" и отправляем инструкцию.
		case "Инструкция":
			userState[message.Chat.ID] = "instruction"
			lastBotMessageID = sendInstruction(bot, message.Chat.ID)
		// При выборе "Тарифы" переключаем состояние на "tariffs" и отправляем описание тарифов.
		case "Тарифы":
			userState[message.Chat.ID] = "tariffs"
			lastBotMessageID = sendTariffs(bot, message.Chat.ID)
		// Если нажата кнопка "Назад в меню", просто отправляем главное меню.
		case "Назад в меню":
			userState[message.Chat.ID] = "main"
			lastBotMessageID = sendMainMenu(bot, message.Chat.ID)
		// Если пользователь отправляет любой другой текст в главном меню, выдаем сообщение об ошибке.
		default:
			lastBotMessageID = sendMessage(bot, message.Chat.ID, "Неизвестная команда. Выберите пункт из меню.")
		}

	// Состояние "question" — пользователь перешёл в режим "Задать вопрос".
	case "question":
		switch message.Text {
		// Если нажата кнопка "Назад в меню", возвращаемся в главное меню.
		case "Назад в меню":
			userState[message.Chat.ID] = "main"
			lastBotMessageID = sendMainMenu(bot, message.Chat.ID)
		// Если введён любой другой текст, считаем его самостоятельным вопросом.
		default:
			if len(message.Text) > 200 {
				// Игнорируем не-текстовые сообщения
				// Отправляем сообщение, что сообщение длинное или не текстовое.
				lastBotMessageID = sendMessage(bot, message.Chat.ID, "Вы отправили слишком длинное сообщение, либо сообщение не текстовое.")
				// После обработки вопроса возвращаем пользователя в главное меню.
				userState[message.Chat.ID] = "main"
			} else {

				userPrompt := `
				Ты профессиональная гадалка-таролог! Разбираешься во всех терминах тарологии, гороскопа и будущего!
				
				Я напишу тебе вопрос касающийся моего будущего.  САМОЕ ГЛАВНОЕ ЕСЛИ МОЙ ВОПРОС НЕ БУДЕТ ПОХОЖ НА ЗАПРОС ПО ГАДАНИЮ, ТО ТЫ ДОЛЖЕН ВЕЖЛИВО ОТКАЗАТЬ В ГАДАНИИ И БОЛЬШЕ НИЧЕГО НЕ ПИСАТЬ!!! САМОЕ ГЛАВНОЕ ЕСЛИ МОЙ ВОПРОС НЕ БУДЕТ ПОХОЖ НА ЗАПРОС ПО ГАДАНИЮ, ТО ТЫ ДОЛЖЕН ВЕЖЛИВО ОТКАЗАТЬ В ГАДАНИИ И БОЛЬШЕ НИЧЕГО НЕ ПИСАТЬ!!!
				
				Твоя задача:
				Сделать мне расклад по трём картам таро. Карты ты должен выбрать сам, назвать и опиши что они значат. Рассказать как пройдёт мой день сегодня исходя из этих карт.
				Расскажи не длинно, не больше 800 символов.
				ОТВЕЧАЙ ТОЛЬКО НА РУССКОМ ЯЗЫКЕ НЕ ИСПОЛЬЗУЙ НИКАКИХ ДРУГИХ ЯЗЫКОВ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ!!!! ОТВЕЧАЙ ТОЛЬКО НА РУССКОМ ЯЗЫКЕ НЕ ИСПОЛЬЗУЙ НИКАКИХ ДРУГИХ ЯЗЫКОВ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ!!!!
				
				Вопрос касающийся моего будущего, ОТВЕЧАЙ ТОЛЬКО НА РУССКОМ ЯЗЫКЕ НЕ ИСПОЛЬЗУЙ НИКАКИХ ДРУГИХ ЯЗЫКОВ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ ОТВЕЧАЙ ТОЛЬКО ИСПОЛЬЗУЯ КИРИЛИЦУ!!!! САМОЕ ГЛАВНОЕ ЕСЛИ МОЙ ВОПРОС НЕ БУДЕТ ПОХОЖ НА ЗАПРОС ПО ГАДАНИЮ, ТО ТЫ ДОЛЖЕН ВЕЖЛИВО ОТКАЗАТЬ В ГАДАНИИ И БОЛЬШЕ НИЧЕГО НЕ ПИСАТЬ!!! :
				` + message.Text

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
				msg := tgbotapi.NewMessage(message.Chat.ID, answer)
				_, err = bot.Send(msg)
				if err != nil {
					log.Printf("Ошибка отправки сообщения: %v", err)
				}

				// Отправляем ответ пользователю.
				lastBotMessageID = sendMessage(bot, message.Chat.ID, answer)
				// После обработки вопроса возвращаем пользователя в главное меню.
				userState[message.Chat.ID] = "main"

			}
		}

	// Состояния "instruction" и "tariffs" — пользователь просматривает информацию.
	case "instruction", "tariffs":
		// В этих режимах единственная допустимая команда — "Назад в меню".
		if message.Text == "Назад в меню" {
			// Переключаем состояние в "main" и отправляем главное меню.
			userState[message.Chat.ID] = "main"
			lastBotMessageID = sendMainMenu(bot, message.Chat.ID)
		} else {
			// Если вводится произвольный текст, выдаем сообщение об ошибке.
			lastBotMessageID = sendMessage(bot, message.Chat.ID, "Неизвестная команда. Выберите пункт 'Назад в меню'.")
		}

	// Если по какой-то причине состояние не соответствует ни одному из вышеописанных, сбрасываем его в "main".
	default:
		userState[message.Chat.ID] = "main"
		lastBotMessageID = sendMainMenu(bot, message.Chat.ID)
	}
}

// Функция sendMessage отправляет текстовое сообщение с клавиатурой, содержащей кнопку "Назад в меню".
// Возвращает ID отправленного сообщения для последующего удаления.
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) int {
	// Создаем сообщение с заданным текстом.
	msg := tgbotapi.NewMessage(chatID, text)
	// Добавляем клавиатуру с единственной кнопкой "Назад в меню".
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{tgbotapi.NewKeyboardButton("Назад в меню")},
		},
		ResizeKeyboard:  true,  // Автоматическая адаптация размеров клавиатуры под устройство пользователя.
		OneTimeKeyboard: false, // Клавиатура исчезает после нажатия на кнопку.
	}
	// Отправляем сообщение и сохраняем его MessageID.
	sentMsg, _ := bot.Send(msg)
	return sentMsg.MessageID
}

// Функция sendMainMenu отправляет главное меню с кнопками для перехода в различные режимы.
// Главное меню содержит кнопки: "Задать вопрос", "Инструкция", "Тарифы" и "Назад в меню".
func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем сообщение с текстом главного меню.
	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	// Определяем клавиатуру главного меню.
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			// Первая строка: кнопка "Задать вопрос"
			{tgbotapi.NewKeyboardButton("Задать вопрос")},
			// Вторая строка: кнопка "Инструкция"
			{tgbotapi.NewKeyboardButton("Инструкция")},
			// Третья строка: кнопка "Тарифы"
			{tgbotapi.NewKeyboardButton("Тарифы")},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
	// Отправляем сообщение и сохраняем его ID.
	sentMsg, _ := bot.Send(msg)
	return sentMsg.MessageID
}

// Функция sendQuestionMenu отправляет меню для режима "Задать вопрос".
// Здесь пользователь может выбрать один из вариантов (с заранее заданными кнопками)
// или ввести свой вопрос вручную (если текст не соответствует кнопкам).
func sendQuestionMenu(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем сообщение с текстом меню вопросов.
	msg := tgbotapi.NewMessage(chatID, "Выберите вопрос или введите его самостоятельно:")
	// Определяем клавиатуру с вариантами и кнопкой "Назад в меню".
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			// Первая строка: кнопка "Что ждёт меня сегодня?"
			{tgbotapi.NewKeyboardButton("Что ждёт меня сегодня?")},
			// Вторая строка: кнопка "Любовный расклад"
			{tgbotapi.NewKeyboardButton("Любовный расклад")},
			// Третья строка: кнопка "Назад в меню"
			{tgbotapi.NewKeyboardButton("Назад в меню")},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
	// Отправляем сообщение и возвращаем его ID.
	sentMsg, _ := bot.Send(msg)
	return sentMsg.MessageID
}

// Функция sendInstruction отправляет сообщение с инструкцией и клавиатурой с кнопкой "Назад в меню".
func sendInstruction(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем сообщение с текстом инструкции.
	msg := tgbotapi.NewMessage(chatID, "Инструкция: \nЗдесь будет ваша инструкция.")
	// Определяем клавиатуру с кнопкой "Назад в меню".
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{tgbotapi.NewKeyboardButton("Назад в меню")},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
	// Отправляем сообщение и возвращаем его ID.
	sentMsg, _ := bot.Send(msg)
	return sentMsg.MessageID
}

// Функция sendTariffs отправляет сообщение с описанием тарифов и клавиатурой с кнопкой "Назад в меню".
func sendTariffs(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем сообщение с текстом тарифов.
	msg := tgbotapi.NewMessage(chatID, "Тарифы: \nЗдесь будет описание тарифов.")
	// Определяем клавиатуру с кнопкой "Назад в меню".
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{tgbotapi.NewKeyboardButton("Назад в меню")},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
	// Отправляем сообщение и возвращаем его ID.
	sentMsg, _ := bot.Send(msg)
	return sentMsg.MessageID
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
