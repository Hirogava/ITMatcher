package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func Request(resume string) (map[string][]string, error) {
	url := "http://localhost:8001/analyze"

	payload := map[string]string{"text": resume}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к AI-сервису: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI-сервис вернул статус: %s", resp.Status)
	}

	var result map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	return result, nil
}
