package main // Основной пакет приложения

import (
	// Стандартная библиотека для логирования
	"fmt"
	log "log"
	"os"

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
		// Если пользователь выбирает один из предложенных вариантов, то:
		case "Что ждёт меня сегодня?":
			// Отправляем заглушку, так как функционал пока не реализован.
			lastBotMessageID = sendMessage(bot, message.Chat.ID, "Функционал пока не реализован.")
		case "Любовный расклад":
			lastBotMessageID = sendMessage(bot, message.Chat.ID, "Функционал пока не реализован.")
		// Если нажата кнопка "Назад в меню", возвращаемся в главное меню.
		case "Назад в меню":
			userState[message.Chat.ID] = "main"
			lastBotMessageID = sendMainMenu(bot, message.Chat.ID)
		// Если введён любой другой текст, считаем его самостоятельным вопросом.
		default:
			// Отправляем сообщение, что вопрос принят. Здесь можно реализовать дополнительную логику по обработке вопроса.
			lastBotMessageID = sendMessage(bot, message.Chat.ID, "Ваш вопрос принят: "+message.Text)
			// После обработки вопроса возвращаем пользователя в главное меню.
			userState[message.Chat.ID] = "main"
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
		ResizeKeyboard:  true, // Автоматическая адаптация размеров клавиатуры под устройство пользователя.
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
