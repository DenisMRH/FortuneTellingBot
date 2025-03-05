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
