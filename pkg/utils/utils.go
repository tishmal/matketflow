package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func ParseEnvInt(envKey string) (int, error) {
	value, err := strconv.Atoi(os.Getenv(envKey))
	if err != nil {
		return 0, fmt.Errorf("invalid %s :%w", envKey, err)
	}
	return value, nil
}

func ValidTime(envKey string) (time.Duration, error) {
	time, err := time.ParseDuration(os.Getenv(envKey))
	if err != nil {
		return 0, fmt.Errorf("invalid %s :%w", envKey, err)
	}
	return time, nil
}

// LoadEnv читает .env файл по указанному пути и записывает KEY=VALUE в os.Environ
func LoadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("Пропущена строка: %s", line)
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)
		os.Setenv(key, val)
	}
	return scanner.Err()
}
