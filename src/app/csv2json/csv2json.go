package csv2json

import (
	"app/utils/logger"
	"app/utils/maputil"
	"fmt"
	"strings"
	"time"

	"github.com/go-gota/gota/dataframe"
)

type (
	CsvData struct {
		DfCsv     *dataframe.DataFrame
		DtUpdated time.Time
	}
)

type Csv2Json interface {
	Process(apiAddress string, queryStrPrm string) *map[string]interface{}
}

type csv2Json struct {
	csvAccessor CsvAccessor
}

func NewCsv2Json(csvAccessorIn CsvAccessor) Csv2Json {
	return &csv2Json{csvAccessor: csvAccessorIn}
}

// 同じCSVデータを何度も読みにいかないためにバックアップしておくための変数
// key	: csv address
var mapCSVDataBackup = make(map[string](*CsvData))

// オープンデータのCSVをJSONに変換する処理
func (c2j *csv2Json) Process(apiAddress string, queryStrPrm string) *map[string]interface{} {
	mapResult := make(map[string]interface{})
	logger.Infos(apiAddress, queryStrPrm)

	dtLastUpdate := time.Date(2000, 1, 1, 1, 1, 0, 0, time.Local)
	hasError := false
	types := strings.Split(queryStrPrm, ",")
	for index, value := range types {
		timeStart := time.Now()

		values := strings.Split(value, ":")
		if len(values) != 2 {
			hasError = true
			message := "invalid query param..."
			logger.Errors(value, message)
			continue
		}

		var mapTmp *map[string]interface{}
		key := values[0]
		apiId := values[1]
		logger.Infof("%d, key=%s, id=%s", index, key, apiId)
		csvData, err := getCSVDataFrame(fmt.Sprintf("%s?id=%s", apiAddress, apiId), c2j.csvAccessor)

		if err != nil {
			hasError = true
			message := "failed to get csv data..."
			logger.Errors(key, message)
			mapTmp = &(map[string]interface{}{key: message})
			logger.Infos(mapTmp)
		} else {
			switch key {
			case "main_summary":
				if mapResult[key] == nil {
					mapTmp = mainSummary(csvData.DfCsv, csvData.DtUpdated)
				} else {
					mapMainSummary := mapResult[key].(map[string]interface{})
					mainSummaryTry2Merge4Deth(csvData.DfCsv, &mapMainSummary)
					mapTmp = nil
				}
			case "patients":
				mapTmp = patients(csvData.DfCsv, csvData.DtUpdated)
			case "patients_summary":
				mapTmp = patientsSummary(csvData.DfCsv, csvData.DtUpdated, c2j.csvAccessor.GetTimeNow())
			case "inspection_persons":
				mapTmp = inspectionPersons(csvData.DfCsv, csvData.DtUpdated)
			case "contacts":
				mapTmp = contacts(csvData.DfCsv, csvData.DtUpdated)
			default:
				hasError = true
				message := "not supported..."
				logger.Errors(key, message)
				mapTmp = &(map[string]interface{}{key: message})
			}
		}

		if mapTmp != nil {
			mapWithKey := map[string]interface{}{key: *mapTmp}
			mapResult = maputil.MergeMaps(mapResult, mapWithKey)
			if csvData.DtUpdated.After(dtLastUpdate) {
				dtLastUpdate = csvData.DtUpdated
			}
		}

		logger.Infof("%s time = %d milliseconds", value, time.Since(timeStart).Milliseconds())
	}

	mapResult["value"] = 0
	mapResult["hasError"] = hasError
	mapResult["lastUpdate"] = dtLastUpdate.Format("2006/01/02 15:04")

	return &mapResult
}

func getCSVDataFrame(apiAddress string, csvAccessor CsvAccessor) (*CsvData, error) {
	data := mapCSVDataBackup[apiAddress]
	var err error
	if data == nil {
		data = &CsvData{}
		data.DfCsv, data.DtUpdated, err = csvAccessor.GetCSVDataFrameFromApi(apiAddress)
		mapCSVDataBackup[apiAddress] = data
	}
	return data, err
}
