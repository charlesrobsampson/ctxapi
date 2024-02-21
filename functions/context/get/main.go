package main

import (
	cntxt "main/functions/context"
	"main/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	userId := request.PathParameters["userId"]
	timestamp, hasTimestamp := request.QueryStringParameters["timestamp"]

	c := cntxt.Context{}

	if hasTimestamp {
		if timestamp == "last" {
			ctx, err := cntxt.GetLastContext(userId)
			if err != nil {
				return utils.HandleError(err)
			}
			c = *ctx
		} else {
			ctx, err := cntxt.GetContext(userId, timestamp)
			if err != nil {
				if err.Error() == "not found" {
					return utils.HandleCode(404, "context not found")
				}
				return utils.HandleError(err)
			}
			c = *ctx
		}
	} else {
		ctx, err := cntxt.GetCurrentContext(userId)
		if err != nil {
			return utils.HandleError(err)
		}
		c = *ctx
	}

	c.NoteString = ""

	contextString, err := c.ToJSONString()

	if err != nil {
		return utils.HandleError(err)
	}

	return utils.HandleSuccess(contextString)
}

func main() {
	lambda.Start(Handler)
}
