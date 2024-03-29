package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	q "main/functions/queue"
	"main/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	userId := request.PathParameters["userId"]

	fmt.Printf("body\n%s\n----\n", []byte(request.Body))
	var body q.Queue
	body.UserId = userId
	err := json.Unmarshal([]byte(request.Body), &body)
	if err != nil {
		dec, err := b64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			panic(fmt.Sprintf("error unmarshaling json body: %s", err))
		}
		err = json.Unmarshal([]byte(dec), &body)
		if err != nil {
			panic(fmt.Sprintf("error unmarshaling json body: %s", err))
		}
	}

	if body.Name == "" {
		return utils.HandleCode(400, "name is required")
	}

	qId, err := body.Update()
	if err != nil {
		return utils.HandleError(err)
	}

	body.Id = qId
	body.NoteString = ""
	output, err := utils.JsonMarshal(body, false)
	// output, err := json.Marshal(body)
	if err != nil {
		return utils.HandleError(err)
	}

	return utils.HandleSuccess(string(output))
}

func main() {
	lambda.Start(Handler)
}
