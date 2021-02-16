package pkgcommon

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"path/filepath"
)

// ResponseBodyParser will parse your api response
// @params body string
// @returns map interface with string keys or error
func ResponseBodyParser(body string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ServeJson serve response in json
// @Params http.ResponseWriter, boolean, string, interface{}, interface{}
func ServeJson(w http.ResponseWriter, status bool, message string, data interface{}, error interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ResponseBuilder(status, message, data, error))
	return
}

// LoadEnvFile load .env file
func LoadEnvFile() {
	err := godotenv.Load(filepath.Join("./", ".env"))
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}
