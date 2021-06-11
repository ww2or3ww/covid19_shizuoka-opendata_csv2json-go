package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"app/csv2json"
	"app/utils/logger"
)

const opendataApiUrl = "https://opendata.pref.shizuoka.jp/api/package_show"

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger.Debugs(request)
	logger.Infof("query = %s\n", request.QueryStringParameters)
	logger.Infof("body  = \n%s\n", request.Body)

	queryStrPrm := request.QueryStringParameters["type"]
	mapData := csv2json.Process(opendataApiUrl, queryStrPrm)

	jsonIndent, err := json.MarshalIndent(mapData, "", "   ")
	if err != nil {
		return createErrorResponse("error occured.", err)
	}

	logger.Debugs(string(jsonIndent))

	return events.APIGatewayProxyResponse{
		Body:       string(jsonIndent),
		StatusCode: 200,
	}, nil
}

func createErrorResponse(message string, err error) (events.APIGatewayProxyResponse, error) {
	logger.Errors(message, err)
	return events.APIGatewayProxyResponse{
		Body:       message,
		StatusCode: 500,
	}, err
}

func init() {
	logger.LogInitialize(logger.Info, 25)
}

func main() {
	logger.Infos("=== START ===")
	lambda.Start(handler)
	logger.Infos("=== COMPLETED ===")
}
