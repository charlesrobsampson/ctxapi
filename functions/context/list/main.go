package main

import (
	"encoding/json"
	cntxt "main/functions/context"
	"main/utils"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	userId := request.PathParameters["userId"]
	start, hasStart := request.QueryStringParameters["start"]
	end, hasEnd := request.QueryStringParameters["end"]
	unit, hasUnit := request.QueryStringParameters["unit"]
	defaultStartBack := 1
	isChangingUnit := false

	yearsBack := 0
	monthsBack := 0

	timeUnits := utils.Unit["h"]
	if hasUnit {
		timeUnits = utils.Unit[unit]
		if unit == "M" || unit == "y" {
			isChangingUnit = true
		}
	}

	if hasStart {
		startBack, err := strconv.Atoi(start)
		if err != nil {
			startBack = defaultStartBack
		}
		if isChangingUnit {
			if unit == "M" {
				monthsBack = startBack
			}
			if unit == "y" {
				yearsBack = startBack
			}
			start = time.Now().UTC().AddDate(-yearsBack, -monthsBack, 0).Format(cntxt.SkDateFormat)
		} else {
			start = time.Now().UTC().Add(-time.Duration(startBack) * timeUnits).Format(cntxt.SkDateFormat)
		}
	} else {
		start = time.Now().UTC().Add(-time.Duration(defaultStartBack) * timeUnits).Format(cntxt.SkDateFormat)
	}
	if hasEnd {
		endBack, err := strconv.Atoi(end)
		if err != nil {
			endBack = 0
		}
		if isChangingUnit {
			if unit == "M" {
				monthsBack = endBack
			}
			if unit == "y" {
				yearsBack = endBack
			}
			end = time.Now().UTC().AddDate(-yearsBack, -monthsBack, 0).Format(cntxt.SkDateFormat)
		} else {
			end = time.Now().UTC().Add(-time.Duration(endBack) * timeUnits).Format(cntxt.SkDateFormat)
		}
	} else {
		end = time.Now().UTC().Format(cntxt.SkDateFormat)
	}

	contexts, err := cntxt.ListContexts(userId, start, end, "")
	if err != nil {
		return utils.HandleError(err)
	}

	list := []cntxt.Context{}

	for _, c := range *contexts {
		c.NoteString = ""
		list = append(list, c)
	}

	ctxJSON, err := json.Marshal(list)
	if err != nil {
		return utils.HandleError(err)
	}

	return utils.HandleSuccess(string(ctxJSON))
}

func main() {
	lambda.Start(Handler)
}
