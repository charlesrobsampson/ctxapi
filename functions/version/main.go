package main

import (
	"fmt"
	"main/utils"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	return utils.HandleSuccess(fmt.Sprintf("v%s", os.Getenv("version")))
}

func main() {
	lambda.Start(Handler)
}
