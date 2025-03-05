package main

import (
	"fmt"
	log "log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Функция handleMessage обрабатывает входящие сообщения от пользователей.
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {

	// // ----------------------- Удаление сообщения пользователя -----------------------
	// // Создаем объект для удаления сообщения пользователя по ID чата и ID сообщения.
	// deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
	// // Отправляем запрос на удаление сообщения пользователя.
	// bot.Send(deleteMsg)

	// // ----------------------- Удаление предыдущего сообщения бота -----------------------
	// // Если переменная lastBotMessageID не равна 0, значит бот уже отправлял сообщение.
	// if lastBotMessageID != 0 {
	// 	// Создаем запрос на удаление предыдущего сообщения бота.
	// 	deleteBotMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, lastBotMessageID)
	// 	// Отправляем запрос на удаление.
	// 	bot.Send(deleteBotMsg)
	// }

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
		// При выборе пункта "🔮 Задать вопрос 🔮" переходим в режим вопроса.
		case "🔮 Задать вопрос 🔮":
			// Устанавливаем состояние для данного чата в "question".
			userState[message.Chat.ID] = "question"
			// Отправляем меню для вопросов, где пользователь может выбрать вариант или ввести вопрос вручную.
			lastBotMessageID = sendQuestionMenu(bot, message.Chat.ID)
		// При выборе "📑 Инструкция 📑" переключаем состояние на "instruction" и отправляем инструкцию.
		case "📑 Инструкция 📑":
			userState[message.Chat.ID] = "instruction"
			lastBotMessageID = sendInstruction(bot, message.Chat.ID)
		// При выборе "💲Тарифы💲" переключаем состояние на "tariffs" и отправляем описание тарифов.
		case "💲Тарифы💲":
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

	// Состояние "question" — пользователь перешёл в режим "🔮 Задать вопрос 🔮".
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
				// Загружаем карты из JSON-файла
				cards, err := loadTarotCards("tarocards.json")
				if err != nil {
					fmt.Println("Ошибка загрузки карт:", err) // Выводим ошибку, если файл не загрузился
					return                                    // Завершаем выполнение программы
				}

				// Выбираем 3 случайные карты
				selected := drawThreeCards(cards)

				// Выводим результат в консоль
				fmt.Println("Ваши карты:")
				cardMsg, cardPrompt := "", ""
				for _, card := range selected { // Итерируемся по выбранным картам
					cardMsg = cardMsg + card.Name + "\n" + card.Description + "\n\n\n"
					cardPrompt = cardPrompt + card.Name + "\n"
				}

				msg := tgbotapi.NewMessage(message.Chat.ID, cardMsg)
				_, err = bot.Send(msg)
				if err != nil {
					log.Printf("Ошибка отправки сообщения: %v", err)
				}

				userPrompt := `
				Ты профессиональная рускоязычная гадалка-таролог! Разбираешься во всех терминах тарологии, во всех картах Таро и их значениях! \n
				Ответь на мой вопросс максимально подробно, опираясь на выпавшие мне карты. \n
				Вопрос:
				` + message.Text + `\n Карты которые мне выпали: \n` + cardMsg

				log.Printf("Сообщение от пользователя: %s", message.Text)

				// Создаём каналы для ответа и ошибок
				answerChan := make(chan string)
				errorChan := make(chan error)

				// Запускаем queryDeepSeek в горутине
				go func() {
					answer, err := queryDeepSeek(userPrompt)
					if err != nil {
						errorChan <- err
						answerChan <- "" // Отправляем пустой ответ
						return
					}
					answerChan <- answer
					errorChan <- nil
				}()

				answer := <-answerChan
				err = <-errorChan

				// Проверяем, что ответ не пустой
				if answer == "" {
					answer = "Извините, я не смог обработать ваш запрос."
				}
				if err != nil {
					answer = "Ошибка при запросе к DeepSeek: " + err.Error()
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
