package main

import (
	q "main/functions/queue"
	"main/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	userId := request.PathParameters["userId"]
	queueId := request.PathParameters["queueId"]
	contextId, ctxPassed := request.QueryStringParameters["contextId"]

	queue, err := q.GetQueue(userId, queueId)
	if err != nil {
		if err.Error() == "not found" {
			return utils.HandleCode(404, "context not found")
		}
		return utils.HandleError(err)
	}

	if ctxPassed {
		queue.ContextId = contextId
	}

	queue, err = queue.Start()
	if err != nil {
		return utils.HandleError(err)
	}

	queue.NoteString = ""

	qString, err := queue.ToJSONString()

	if err != nil {
		return utils.HandleError(err)
	}

	return utils.HandleSuccess(qString)
}

func main() {
	lambda.Start(Handler)
}
