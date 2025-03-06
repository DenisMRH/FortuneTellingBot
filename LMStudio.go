package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type SaigaRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type Choice struct {
	Text string `json:"text"`
}

type SaigaResponse struct {
	Choices []Choice `json:"choices"`
}

func callSaigaMistral(prompt string) (string, error) {
	url := "http://localhost:1234/v1/completions" // URL API LM Studio
	data := SaigaRequest{
		Model:     "gigachat-20b-a3b-instruct",
		Prompt:    prompt,
		MaxTokens: 2000, // Ограничиваем длину ответа
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var saigaResp SaigaResponse
	if err := json.NewDecoder(resp.Body).Decode(&saigaResp); err != nil {
		return "", err
	}

	if len(saigaResp.Choices) == 0 {
		return "Ошибка: пустой ответ от модели", nil
	}

	return saigaResp.Choices[0].Text, nil
}

func ruToEn(ruText string) (enText string) {
	if ruText == "" {
		ruText = "Ошибка: модель вернула пустой ответ"
	}
	enText, err := callSaigaMistral(`Ты — древняя и мудрая гадалка, которая общается с человеком при помощи карт Таро.
Твоё знание безгранично, но ты передаёшь его в завуалированной, мистической форме,
словно старая провидица из далёких времён.

**Твой стиль ответа**:
- Говори только на русском языке.
- Излагай мысли загадочно, таинственно, но при этом понятно и доступно человеку.
- Строй речь так, чтобы читатель погружался в атмосферу магического ритуала.

**Контекст**:
- Тебе задали вопрос: ⏰ Что ждёт меня сегодня? ⏰
- Выпали следующие карты Таро (с краткими описаниями):
  🃏Правосудие (прямое положение)
Правосудие символизирует справедливость, закон и карму. Это карта баланса, которая указывает на необходимость принимать ответственность за свои действия. Она призывает к честности и объективности в принятии решений.


🃏Туз Кубков (перевёрнутое положение)
В перевернутом виде Туз Кубков может указывать на эмоциональную пустоту, разочарование или блокировку чувств. Возможны трудности с принятием любви или творческим кризис. Это предупреждение о необходимости исцелить свое сердце и восстановить связь с эмоциями.


🃏Рыцарь Жезлов (прямое положение)
Рыцарь Жезлов символизирует действие, смелость и амбиции. Это карта движения вперед, которая призывает к решительности.


**Твоя задача**:
1. Перечитай вопрос.
2. Проинтерпретируй три карты Таро в контексте данного вопроса.
3. Отвечай так, словно ты обладаешь тайными знаниями и передаёшь их клиенту, слегка приоткрывая завесу будущего.
4. Будь убедительна и опирайся на значения карт.

**Формат ответа**:
- Текстовое сообщение в стиле мистической гадалки.
- Избегай технических подробностей, говори образно и символично.
- Обращайся к собеседнику на «ты» или «вы», если уместно, но соблюдай тон волшебницы.
` + ruText)
	if err != nil {
		log.Printf("Ошибка запроса к модели: ")
	}

	if enText == "" {
		log.Printf("Ошибка: модель вернула пустой ответ")
	}

	return

}
func enToRu(enText string) (ruText string) {
	if enText == "" {
		enText = "Ошибка: модель вернула пустой ответ"
	}
	ruText, err := callSaigaMistral(`Переведи это на русский:` + enText)
	if err != nil {
		log.Printf("Ошибка запроса к модели: ")
	}

	if enText == "" {
		log.Printf("Ошибка: модель вернула пустой ответ")
	}

	return

}
