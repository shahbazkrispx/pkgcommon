package pkgcommon

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// MessageAttributesBodyParser will parse aws message attributes
// @params body string
// @returns map string with string keys or error
// MessageAttributesBodyParser will parse aws message attributes more efficiently
// @params body string
// @returns map string with string keys or error
func MessageAttributesBodyParser(msgBody string) (map[string]string, error) {
	// Pre-allocate result map
	res := make(map[string]string)

	// Parse JSON directly into a map structure
	var data struct {
		MessageAttributes map[string]struct {
			Value interface{} `json:"Value"`
		} `json:"MessageAttributes"`
	}

	if err := json.Unmarshal([]byte(msgBody), &data); err != nil {
		return nil, err
	}

	// Extract values directly into result map
	for k, attr := range data.MessageAttributes {
		res[k] = fmt.Sprint(attr.Value)
	}

	return res, nil
}

// func MessageAttributesBodyParser(msgBody string) (map[string]string, error) {
// 	var data map[string]interface{}
// 	res := make(map[string]string)

// 	err := json.Unmarshal([]byte(msgBody), &data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for k, d2 := range data["MessageAttributes"].(map[string]interface{}) {
// 		v := fmt.Sprintf("%v", d2.(map[string]interface{})["Value"])
// 		res[k] = v
// 	}
// 	return res, nil
// }

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
}

// LoadEnvFile load .env file
func LoadEnvFile() {
	err := godotenv.Load(filepath.Join("./", ".env"))
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

func GetQueueURL(queue string) string {
	LoadEnvFile()
	if os.Getenv("APP_ENV") == "prod" || os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "" {
		return fmt.Sprintf("https://sqs.%s.amazonaws.com/%s/%s", os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCOUNT_ID"), queue)
	} else {
		return fmt.Sprintf("https://sqs.%s.amazonaws.com/%s/%s_%s", os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCOUNT_ID"), os.Getenv("APP_ENV"), queue)
	}
}

func GetSNSArn(sns string) string {
	LoadEnvFile()
	if os.Getenv("APP_ENV") == "prod" || os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "" {
		return fmt.Sprintf("arn:aws:sns:%s:%s:%s", os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCOUNT_ID"), sns)
	} else {
		return fmt.Sprintf("arn:aws:sns:%s:%s:%s_%s", os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCOUNT_ID"), os.Getenv("APP_ENV"), sns)
	}
}

func MakeSNSArn(sns string) string {
	LoadEnvFile()
	return fmt.Sprintf("arn:aws:sns:%s:%s:%s", os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCOUNT_ID"), sns)
}
