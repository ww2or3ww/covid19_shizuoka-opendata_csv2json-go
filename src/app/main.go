package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"app/csv2json"
	"app/utils/logger"
)

const defaultTypes = "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,main_summary:92f9ebcd-a3f1-4d5d-899b-d69214294a45,patients:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,patients_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,inspection_persons:d4827176-d887-412a-9344-f84f161786a2,contacts:1b57f2c0-081e-4664-ba28-9cce56d0b314"

const opendataApiUrl = "https://opendata.pref.shizuoka.jp/api/package_show"

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger.Debugs(request)
	logger.Infof("query = %s\n", request.QueryStringParameters)
	logger.Infof("body  = \n%s\n", request.Body)

	queryStrPrm := request.QueryStringParameters["type"]
	if queryStrPrm == "" {
		queryStrPrm = defaultTypes
	}
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
