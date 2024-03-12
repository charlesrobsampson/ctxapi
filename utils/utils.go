package utils

import (
	"bytes"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

func HandleError(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: 500,
	}, err
}

func HandleSuccess(message string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       message,
		StatusCode: 200,
	}, nil
}

func HandleCode(code int, message string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       message,
		StatusCode: code,
	}, nil
}

func IsNullJSON(m json.RawMessage) bool {
	return len(m) == 0 || string(m) == "null"
}

func JsonMarshal(data interface{}, escapeHTML bool) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(escapeHTML)
	err := encoder.Encode(data)
	return buffer.Bytes(), err
}
