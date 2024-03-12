package main

import (
	q "main/functions/queue"
	"main/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	userId := request.PathParameters["userId"]
	listAll, allSent := request.QueryStringParameters["all"]

	pendingOnly := true

	if allSent && listAll == "true" {
		pendingOnly = false
	}

	queue, err := q.ListQueue(userId, pendingOnly)
	if err != nil {
		if err.Error() == "not found" {
			return utils.HandleCode(404, "context not found")
		}
		return utils.HandleError(err)
	}

	list := []q.Queue{}

	for _, qu := range *queue {
		qu.NoteString = ""
		list = append(list, qu)
	}
	// contextString, err := queue.ToJSONString()

	qJSON, err := utils.JsonMarshal(list, false)
	// qJSON, err := json.Marshal(list)
	if err != nil {
		return utils.HandleError(err)
	}

	// if err != nil {
	// 	return utils.HandleError(err)
	// }

	return utils.HandleSuccess(string(qJSON))
}

func main() {
	lambda.Start(Handler)
}
