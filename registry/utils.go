package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	M int `json:"bits"`
}

func ReadFromConfig() (int, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	filePath := filepath.Join(currentDir, "config.json")

	file, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return 0, err
	}

	return config.M, nil
}

func ObtainAddress(node string) string {

	parts := strings.Split(node, ":")
	addr := "0.0.0.0:" + parts[1]

	return addr

}
