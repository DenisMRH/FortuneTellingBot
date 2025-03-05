package main

import (
	// Библиотека для работы с Telegram Bot API
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
// Главное меню содержит кнопки: "🔮 Задать вопрос 🔮", "📑 Инструкция 📑", "💲Тарифы💲" и "Назад в меню".
func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем сообщение с текстом главного меню.
	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	// Определяем клавиатуру главного меню.
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			// Первая строка: кнопка "🔮 Задать вопрос 🔮"
			{tgbotapi.NewKeyboardButton("🔮 Задать вопрос 🔮")},
			// Вторая строка: кнопка "📑 Инструкция 📑"
			{tgbotapi.NewKeyboardButton("📑 Инструкция 📑")},
			// Третья строка: кнопка "💲Тарифы💲"
			{tgbotapi.NewKeyboardButton("💲Тарифы💲")},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
	// Отправляем сообщение и сохраняем его ID.
	sentMsg, _ := bot.Send(msg)
	return sentMsg.MessageID
}

// Функция sendQuestionMenu отправляет меню для режима "🔮 Задать вопрос 🔮".
// Здесь пользователь может выбрать один из вариантов (с заранее заданными кнопками)
// или ввести свой вопрос вручную (если текст не соответствует кнопкам).
func sendQuestionMenu(bot *tgbotapi.BotAPI, chatID int64) int {
	// Создаем сообщение с текстом меню вопросов.
	msg := tgbotapi.NewMessage(chatID, "Выберите вопрос или введите его самостоятельно:")
	// Определяем клавиатуру с вариантами и кнопкой "Назад в меню".
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			// Первая строка: кнопка "Что ждёт меня сегодня?"
			{tgbotapi.NewKeyboardButton("⏰ Что ждёт меня сегодня? ⏰")},
			// Вторая строка: кнопка "💔 Любовный расклад 💔"
			{tgbotapi.NewKeyboardButton("💔 Любовный расклад 💔")},
			// Вторая строка: кнопка "👩🏻‍💼 Карьерный расклад 👩🏻‍💼"
			{tgbotapi.NewKeyboardButton("👩🏻‍💼 Карьерный расклад 👩🏻‍💼")},
			// Вторая строка: кнопка "💵 Финансовый расклад 💵"
			{tgbotapi.NewKeyboardButton("💵 Финансовый расклад 💵")},
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
	msg := tgbotapi.NewMessage(chatID, "📑 Инструкция 📑: \nЗдесь будет наша инструкция.")
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
	msg := tgbotapi.NewMessage(chatID, "💲Тарифы💲: \nЗдесь будет описание тарифов.")
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
