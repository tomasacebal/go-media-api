package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func loadDotEnvFiles() error {
	paths := []string{".env"}
	executablePath, err := os.Executable()
	if err == nil {
		executableDotEnv := filepath.Join(filepath.Dir(executablePath), ".env")
		if !samePath(".env", executableDotEnv) {
			paths = append(paths, executableDotEnv)
		}
	}

	for _, path := range paths {
		if err := loadDotEnv(path); err != nil {
			return err
		}
	}
	return nil
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("leer %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := applyDotEnvLine(line, lineNumber); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("leer %s: %w", path, err)
	}
	return nil
}

func applyDotEnvLine(line string, lineNumber int) error {
	line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return fmt.Errorf(".env linea %d invalida", lineNumber)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf(".env linea %d sin clave", lineNumber)
	}
	if _, exists := os.LookupEnv(key); exists {
		return nil
	}

	parsedValue, err := parseDotEnvValue(strings.TrimSpace(value), lineNumber)
	if err != nil {
		return err
	}
	if err := os.Setenv(key, parsedValue); err != nil {
		return fmt.Errorf("setear %s desde .env: %w", key, err)
	}
	return nil
}

func parseDotEnvValue(value string, lineNumber int) (string, error) {
	if value == "" {
		return "", nil
	}
	if strings.HasPrefix(value, `"`) {
		parsedValue, err := strconv.Unquote(value)
		if err != nil {
			return "", fmt.Errorf(".env linea %d tiene comillas invalidas", lineNumber)
		}
		return parsedValue, nil
	}
	if strings.HasPrefix(value, "'") {
		if !strings.HasSuffix(value, "'") || len(value) == 1 {
			return "", fmt.Errorf(".env linea %d tiene comillas invalidas", lineNumber)
		}
		return strings.TrimSuffix(strings.TrimPrefix(value, "'"), "'"), nil
	}
	if commentIndex := strings.Index(value, " #"); commentIndex >= 0 {
		value = value[:commentIndex]
	}
	return strings.TrimSpace(value), nil
}

func samePath(left string, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return filepath.Clean(left) == filepath.Clean(right)
	}
	return strings.EqualFold(filepath.Clean(leftAbs), filepath.Clean(rightAbs))
}
