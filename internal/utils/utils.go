package utils

import (
	"encoding/json"
	"log"
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

const projectDirName = "terraform-provider-paralus"

func MultiEnvSearch(ks []string) string {
	for _, k := range ks {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// Load an environment variable
func LoadEnv() {
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))

	err := godotenv.Load(string(rootPath) + `/.env`)

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

// Convert a json to a string map interface
func JsonToMap(jsonStr string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Convert a map to json string
func MapToJsonString(jsonMap map[string]string) (string, error) {
	jsonBytes, err := json.Marshal(&jsonMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
