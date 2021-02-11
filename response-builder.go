package pkgcommon

type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// ResponseBuilder build a standard api response
// @Param	status	bool, msg string, data interface{}, errors interface{}
// @Returns Response type
func ResponseBuilder(status bool, msg string, data interface{}, errors interface{}) Response {
	return Response{
		Status:  status,
		Message: msg,
		Data:    data,
		Error:   errors,
	}
}
