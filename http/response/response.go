package response

import "encoding/json"

// Response is the result to user and it will be convert to a json object
type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

// createResponse function convert object to json response
func createResponse(result Response) []byte {
	res, _ := json.Marshal(result)
	return res
}

// Success function return a successful message to user
func Success(result interface{}) []byte {
	return createResponse(Response{
		StatusCode: 200,
		Message:    "ok",
		Data:       result,
	})
}

// Failed function return a failed message to user
func Failed(message string) []byte {
	return createResponse(Response{
		StatusCode: 500,
		Message:    message,
	})
}
