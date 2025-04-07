package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func Request(resume string) (map[string][]string, error) {
	cmd := exec.Command("python3.12", "services/ai/main.py")
	cmd.Stdin = bytes.NewBufferString(resume)
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %v, output: %s", err, string(output))
	}

	var result map[string][]string
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("Ошибка декодирования json: %v", err)
	}

	return result, nil
}
