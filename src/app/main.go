package main

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"app/csv2json"
	"app/utils/logger"
)

// クエリパラメータが無かった場合のデフォルト(浜松市)
const defaultTypes = "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,main_summary:92f9ebcd-a3f1-4d5d-899b-d69214294a45,patients:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,patients_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,inspection_persons:d4827176-d887-412a-9344-f84f161786a2,contacts:1b57f2c0-081e-4664-ba28-9cce56d0b314"

// APIアドレス(ふじの国オープンデータカタログ)
const opendataApiUrl = "https://opendata.pref.shizuoka.jp/api/package_show"

type Csv2Json interface {
	Process(apiAddress string, queryStrPrm string) (*csv2json.Result, error)
}

var c2j Csv2Json

// AWS Lambda エンドポイント
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	timeStart := time.Now()
	logger.Debugs(request)
	logger.Infof("query = %s\n", request.QueryStringParameters)
	logger.Infof("body  = \n%s\n", request.Body)

	// クエリパラメータ取得
	queryStrPrm := request.QueryStringParameters["type"]
	if queryStrPrm == "" {
		queryStrPrm = defaultTypes
	}

	// csv2json
	mapData, err := c2j.Process(opendataApiUrl, queryStrPrm)
	if err != nil {
		return createErrorResponse("convert csv into json failed", err)
	}

	// mapをインデント付きのJSONに整形してBodyとして返す
	jsonIndent, err := json.MarshalIndent(mapData, "", "   ")
	if err != nil {
		return createErrorResponse("error occured.", err)
	}

	logger.Debugs(string(jsonIndent))
	logger.Infof("total time = %d milliseconds", time.Since(timeStart).Milliseconds())

	return events.APIGatewayProxyResponse{
		Body:       string(jsonIndent),
		StatusCode: 200,
	}, nil
}

// エラー時のレスポンス
func createErrorResponse(message string, err error) (events.APIGatewayProxyResponse, error) {
	logger.Errors(message, err)
	return events.APIGatewayProxyResponse{
		Body:       message,
		StatusCode: 500,
	}, err
}

// mainメソッドの前に呼ばれる初期化処理
func init() {
	// LogLvを環境変数から取得してLog初期設定する
	logLv := logger.Error
	envLogLv := os.Getenv("LOG_LEVEL")
	logger.Infos(envLogLv)
	if envLogLv != "" {
		n, _ := strconv.Atoi(envLogLv)
		logLv = logger.LogLv(n)
	}
	logger.LogInitialize(logLv, 25)

	// 本番用のCSV2JSONをDIしておく
	c2j = csv2json.NewCsv2Json(csv2json.NewCsvAccessor())
}

// アプリケーションエンドポイント
func main() {
	logger.Infos("=== START ===")
	lambda.Start(handler)
	logger.Infos("=== COMPLETED ===")
}
