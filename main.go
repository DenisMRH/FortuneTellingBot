package main // Объявление пакета main — точка входа в программу

// Импорт необходимых библиотек:
import (
	// Стандартная библиотека log для логирования ошибок и информационных сообщений
	log "log"
	// Библиотека для работы с Telegram Bot API
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Глобальная переменная для хранения ID последнего сообщения, отправленного ботом.
// Используется для удаления предыдущего ответа бота перед отправкой нового.
var lastBotMessageID int

// Функция main — точка входа программы
func main() {
	// Создание нового экземпляра бота, используя ваш уникальный токен.
	// Функция NewBotAPI возвращает объект bot и ошибку (err)
	bot, err := tgbotapi.NewBotAPI("5318879758:AAHu_iTb1iY_s6Go5kCkabKQnDSNHZIATt8")
	// Если произошла ошибка (например, неверный токен), программа выводит ошибку и завершается
	if err != nil {
		log.Panic(err)
	}

	// Включаем режим отладки, чтобы получать подробный лог работы бота
	bot.Debug = true

	// Выводим в лог сообщение об успешной авторизации бота с указанием его имени
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Создаем объект конфигурации для получения обновлений (новых сообщений)
	// Передаем 0, что означает, что мы хотим получать все обновления начиная с самого первого
	updateConfig := tgbotapi.NewUpdate(0)
	// Устанавливаем таймаут для получения обновлений. Если за 60 секунд ничего не придет — соединение разорвается.
	updateConfig.Timeout = 60

	// Получаем канал, по которому будут приходить обновления от Telegram
	updates := bot.GetUpdatesChan(updateConfig)

	// Бесконечный цикл для обработки каждого обновления, полученного из канала updates
	for update := range updates {
		// Если обновление содержит сообщение (а не, например, callback-запрос), то:
		if update.Message != nil {
			// Передаем сообщение в функцию обработки handleMessage, где будет осуществлена логика ответа
			handleMessage(bot, update.Message)
		}
	}
}

// Функция handleMessage обрабатывает входящие сообщения от пользователей.
// Принимает два аргумента: объект бота и сообщение от пользователя.
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// ----------------------- Удаление сообщения пользователя -----------------------
	// Создаем запрос на удаление сообщения пользователя, используя ID чата и ID самого сообщения.
	deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
	// Отправляем запрос на удаление. Это нужно, чтобы в чате оставался только ответ бота.
	bot.Send(deleteMsg)

	// ----------------------- Удаление предыдущего сообщения бота -----------------------
	// Если переменная lastBotMessageID не равна 0, значит бот ранее отправлял сообщение.
	if lastBotMessageID != 0 {
		// Создаем запрос на удаление предыдущего сообщения бота по сохраненному ID.
		deleteBotMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, lastBotMessageID)
		// Отправляем запрос на удаление предыдущего сообщения бота.
		bot.Send(deleteBotMsg)
	}

	// ----------------------- Обработка текста входящего сообщения -----------------------
	// Используем конструкцию switch для определения, какое действие выполнить в зависимости от текста сообщения.
	switch message.Text {
	// Если пользователь отправил "/start", запускается главное меню.
	case "/start":
		// Вызывается функция sendMainMenu, которая отправляет главное меню.
		// Возвращается ID отправленного сообщения, который сохраняется в переменной lastBotMessageID.
		lastBotMessageID = sendMainMenu(bot, message.Chat.ID)
	// Если пользователь выбрал "Задать вопрос", отправляется меню с вопросами.
	case "Задать вопрос":
		lastBotMessageID = sendQuestionMenu(bot, message.Chat.ID)
	// Если пользователь выбрал "Инструкция", отправляется сообщение с инструкцией.
	case "Инструкция":
		lastBotMessageID = sendInstruction(bot, message.Chat.ID)
	// Если пользователь выбрал "Тарифы", отправляется сообщение с описанием тарифов.
	case "Тарифы":
		lastBotMessageID = sendTariffs(bot, message.Chat.ID)
	// Если пользователь выбрал один из вариантов вопроса ("Что ждёт меня сегодня?" или "Любовный расклад"),
	// отправляется сообщение-заглушка, так как функционал ещё не реализован.
	case "Что ждёт меня сегодня?", "Любовный расклад":
		lastBotMessageID = sendMessage(bot, message.Chat.ID, "Функционал пока не реализован.")
	// Если пользователь нажал "Назад в меню", отправляется главное меню.
	case "Назад в меню":
		lastBotMessageID = sendMainMenu(bot, message.Chat.ID)
	// Если текст сообщения не соответствует ни одному из вариантов, отправляется сообщение об ошибке.
	default:
		lastBotMessageID = sendMessage(bot, message.Chat.ID, "Неизвестная команда. Выберите пункт из меню.")
	}
}

// Функция sendMessage отправляет произвольное текстовое сообщение с кнопкой "Назад в меню".
// Она принимает объект бота, ID чата и текст сообщения, а возвращает ID отправленного сообщения.
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) int {
	// Создаем новое сообщение с заданным текстом для указанного чата.
	msg := tgbotapi.NewMessage(chatID, text)
	// Устанавливаем клавиатуру с единственной кнопкой "Назад в меню".
	// Это гарантирует, что у пользователя всегда будет возможность вернуться в главное меню.
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			// Создаем ряд клавиатуры с одной кнопкой "Назад в меню"
			{tgbotapi.NewKeyboardButton("Назад в меню")},
		},
		ResizeKeyboard:  true,  // Автоматически подгоняет размер клавиатуры под экран пользователя
		OneTimeKeyboard: false, // Клавиатура исчезает после нажатия на кнопку
	}
	// Отправляем сообщение и сохраняем объект отправленного сообщения (sentMsg).
	sentMsg, _ := bot.Send(msg)
	// Возвращаем MessageID отправленного сообщения для последующего удаления.
	return sentMsg.MessageID
}

// Функция sendMainMenu отправляет главное меню.
// Главное меню содержит кнопки для основных действий, включая "Назад в меню" для единообразия.
func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем новое сообщение с текстом "Выберите действие:" для пользователя.
	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	// Определяем клавиатуру, которая будет показана пользователю вместе с сообщением.
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			// Первая строка: кнопка "Задать вопрос"
			{tgbotapi.NewKeyboardButton("Задать вопрос")},
			// Вторая строка: кнопка "Инструкция"
			{tgbotapi.NewKeyboardButton("Инструкция")},
			// Третья строка: кнопка "Тарифы"
			{tgbotapi.NewKeyboardButton("Тарифы")},
		},
		ResizeKeyboard:  true,  // Клавиатура адаптируется под размер экрана
		OneTimeKeyboard: false, // Клавиатура исчезает после выбора кнопки
	}
	// Отправляем сообщение и сохраняем ID отправленного сообщения.
	sentMsg, _ := bot.Send(msg)
	// Возвращаем MessageID для возможности последующего удаления.
	return sentMsg.MessageID
}

// Функция sendQuestionMenu отправляет меню с вариантами вопросов.
// Меню включает опции вопросов и всегда содержит кнопку "Назад в меню".
func sendQuestionMenu(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем сообщение с текстом "Выберите вопрос:".
	msg := tgbotapi.NewMessage(chatID, "Выберите вопрос:")
	// Определяем клавиатуру с тремя кнопками: два варианта вопроса и кнопка для возврата в главное меню.
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			// Первая строка: кнопка "Что ждёт меня сегодня?"
			{tgbotapi.NewKeyboardButton("Что ждёт меня сегодня?")},
			// Вторая строка: кнопка "Любовный расклад"
			{tgbotapi.NewKeyboardButton("Любовный расклад")},
			// Третья строка: кнопка "Назад в меню" для возврата
			{tgbotapi.NewKeyboardButton("Назад в меню")},
		},
		ResizeKeyboard:  true,  // Автоматическая подгонка размеров клавиатуры
		OneTimeKeyboard: false, // Клавиатура исчезает после выбора
	}
	// Отправляем сообщение и сохраняем его MessageID.
	sentMsg, _ := bot.Send(msg)
	// Возвращаем MessageID для удаления при следующем обновлении.
	return sentMsg.MessageID
}

// Функция sendInstruction отправляет сообщение с инструкцией для пользователя.
// Сообщение содержит текст-инструкцию и всегда включает кнопку "Назад в меню".
func sendInstruction(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем сообщение с текстом инструкции.
	msg := tgbotapi.NewMessage(chatID, "Инструкция: \nЗдесь будет ваша инструкция.")
	// Определяем клавиатуру с единственной кнопкой "Назад в меню".
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{tgbotapi.NewKeyboardButton("Назад в меню")},
		},
		ResizeKeyboard:  true,  // Адаптация клавиатуры под экран
		OneTimeKeyboard: false, // Клавиатура скрывается после выбора
	}
	// Отправляем сообщение и сохраняем его MessageID.
	sentMsg, _ := bot.Send(msg)
	// Возвращаем MessageID отправленного сообщения.
	return sentMsg.MessageID
}

// Функция sendTariffs отправляет сообщение с описанием тарифов.
// Также включает клавиатуру с кнопкой "Назад в меню".
func sendTariffs(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем сообщение с текстом тарифов.
	msg := tgbotapi.NewMessage(chatID, "Тарифы: \nЗдесь будет описание тарифов.")
	// Определяем клавиатуру с кнопкой "Назад в меню".
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{tgbotapi.NewKeyboardButton("Назад в меню")},
		},
		ResizeKeyboard:  true,  // Подгонка клавиатуры под размер экрана
		OneTimeKeyboard: false, // Клавиатура скрывается после нажатия
	}
	// Отправляем сообщение и сохраняем его MessageID.
	sentMsg, _ := bot.Send(msg)
	// Возвращаем MessageID для последующего удаления, если потребуется.
	return sentMsg.MessageID
}
