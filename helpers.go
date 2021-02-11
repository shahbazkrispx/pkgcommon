package pkgcommon

import "encoding/json"

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
