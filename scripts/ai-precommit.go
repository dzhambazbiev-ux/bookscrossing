package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

type ReviewResponse struct {
	Status   string    `json:"status"`
	Warnings []Warning `json:"warnings"`
}

type Warning struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
}

func main() {
	// Загружаем .env файл перед проверкой переменных
	if err := godotenv.Load(".env"); err != nil {
		// Игнорируем ошибку, если файл не найден (переменные могут быть установлены в системе)
	}

	// Получаем diff измененных Go файлов
	diff, err := getStagedGoDiff()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения diff: %v\n", err)
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println("Нет измененных Go файлов для проверки")
		os.Exit(0)
	}

	// Читаем промпт
	promptPath := ".ai/go-precommit-prompt.md"
	prompt, err := readFile(promptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения промпта: %v\n", err)
		os.Exit(1)
	}

	// Формируем полный запрос
	fullPrompt := prompt + "\n\nDIFF:\n" + diff

	// Отправляем запрос к OpenAI
	response, err := callOpenAI(fullPrompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка вызова OpenAI: %v\n", err)
		os.Exit(1)
	}

	// Парсим JSON ответ
	var review ReviewResponse
	if err := json.Unmarshal([]byte(response), &review); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка парсинга ответа OpenAI: %v\n", err)
		fmt.Fprintf(os.Stderr, "Ответ: %s\n", response)
		os.Exit(1)
	}

	// Выводим результаты
	if review.Status == "warning" && len(review.Warnings) > 0 {
		fmt.Println("\n❌ Найдены проблемы в коде:\n")
		for _, w := range review.Warnings {
			fmt.Printf("  📍 %s:%d\n", w.File, w.Line)
			fmt.Printf("     %s\n\n", w.Message)
		}
		fmt.Println("Исправьте ошибки перед коммитом.")
		os.Exit(1)
	}

	fmt.Println("✅ Код прошел проверку")
	os.Exit(0)
}

func getStagedGoDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--", "*.go")
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Git diff возвращает код 1, если нет изменений
			return "", nil
		}
		return "", err
	}
	return string(output), nil
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func callOpenAI(prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY не установлен. Установите переменную окружения")
	}

	// Определяем модель OpenAI
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini" // модель по умолчанию
	}

	// Формируем JSON запрос
	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Отправляем HTTP запрос к OpenAI API
	cmd := exec.Command("curl", "-s",
		"-H", "Content-Type: application/json",
		"-H", fmt.Sprintf("Authorization: Bearer %s", apiKey),
		"-d", string(jsonData),
		"https://api.openai.com/v1/chat/completions")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ошибка curl: %v, stderr: %s", err, stderr.String())
	}

	// Парсим ответ OpenAI API
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &apiResponse); err != nil {
		// Выводим сырой ответ для отладки
		return "", fmt.Errorf("ошибка парсинга ответа API: %v\nСырой ответ: %s", err, stdout.String())
	}

	if apiResponse.Error.Message != "" {
		return "", fmt.Errorf("ошибка API: %s", apiResponse.Error.Message)
	}

	if len(apiResponse.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от API. Сырой ответ: %s", stdout.String())
	}

	return apiResponse.Choices[0].Message.Content, nil
}
