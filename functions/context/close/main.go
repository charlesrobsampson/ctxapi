package main

import (
	"fmt"
	cntxt "main/functions/context"
	"main/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	userId := request.PathParameters["userId"]
	timestamp := request.PathParameters["contextId"]

	contextId := fmt.Sprintf("context#%s", timestamp)

	currentContext, err := cntxt.GetCurrentContext(userId)
	fmt.Printf("close GetCurrentContext err: %+v\n----\n", err)
	if err != nil {
		return utils.HandleError(err)
	}

	if timestamp == "current" {
		if currentContext.Pk == "" {
			return utils.HandleCode(404, "no current context")
		}
		contextId = currentContext.Sk
	}
	fmt.Printf("contextId vs currentContext: '%s'\nvs\n'%+vs'\n----\n", contextId, currentContext)

	if contextId == currentContext.Sk {
		fmt.Printf("context '%s' is current context\n", contextId)
		cntxt.SetLastContext(userId, contextId)
		cntxt.SetCurrentContext(userId, "")
		err := currentContext.Close()
		if err != nil {
			return utils.HandleError(err)
		}
		return utils.HandleSuccess(fmt.Sprintf("context '%s' closed", currentContext.Name))
	}
	c, err := cntxt.GetContext(userId, contextId)
	fmt.Printf("close GetContext err: %+v\n----\n", err)
	if err.Error() == "not found" {
		return utils.HandleCode(404, fmt.Sprintf("context '%s' not found", contextId))
	}
	if err != nil {
		return utils.HandleError(err)
	}

	err = c.Close()

	if err != nil {
		return utils.HandleError(err)
	}

	return utils.HandleSuccess(fmt.Sprintf("context '%s' closed", c.Name))
}

func main() {
	lambda.Start(Handler)
}
