package csv2json

import (
	"app/utils/logger"
	"app/utils/maputil"
	"errors"
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
	Process(apiAddress string, queryStrPrm string) (*map[string]interface{}, error)
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
func (c2j *csv2Json) Process(apiAddress string, queryStrPrm string) (*map[string]interface{}, error) {
	mapResult := make(map[string]interface{})
	logger.Infos(apiAddress, queryStrPrm)

	dtLastUpdate := time.Date(2000, 1, 1, 1, 1, 0, 0, time.Local)
	types := strings.Split(queryStrPrm, ",")
	for index, value := range types {
		timeStart := time.Now()

		values := strings.Split(value, ":")
		if len(values) != 2 {
			message := "invalid query param..."
			logger.Errors(value, message)
			return nil, errors.New("invalid query param")
		}

		var mapTmp *map[string]interface{}
		key := values[0]
		apiId := values[1]
		logger.Infof("%d, key=%s, id=%s", index, key, apiId)
		csvData, err := getCSVDataFrame(fmt.Sprintf("%s?id=%s", apiAddress, apiId), c2j.csvAccessor)

		if err != nil {
			message := "failed to get csv data..."
			logger.Errors(key, message)
			return nil, errors.New("failed to get cav data")
		}

		switch key {
		case "main_summary":
			if mapResult[key] == nil {
				mapTmp, err = mainSummary(csvData.DfCsv, csvData.DtUpdated)
				if err != nil {
					return nil, err
				}
			} else {
				mapMainSummary, ok := mapResult[key].(map[string]interface{})
				if !ok {
					return nil, err
				}
				err = mainSummaryTry2Merge4Deth(csvData.DfCsv, &mapMainSummary)
				if err != nil {
					return nil, err
				}
				mapTmp = nil
			}
		case "patients":
			mapTmp, err = patients(csvData.DfCsv, csvData.DtUpdated)
			if err != nil {
				return nil, err
			}
		case "patients_summary":
			mapTmp, err = patientsSummary(csvData.DfCsv, csvData.DtUpdated, c2j.csvAccessor.GetTimeNow())
			if err != nil {
				return nil, err
			}
		case "inspection_persons":
			mapTmp, err = inspectionPersons(csvData.DfCsv, csvData.DtUpdated)
			if err != nil {
				return nil, err
			}
		case "contacts":
			mapTmp, err = contacts(csvData.DfCsv, csvData.DtUpdated)
			if err != nil {
				return nil, err
			}
		default:
			message := "not supported..."
			logger.Errors(key, message)
			return nil, errors.New("not supported")
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
	mapResult["hasError"] = false
	mapResult["lastUpdate"] = dtLastUpdate.Format("2006/01/02 15:04")

	return &mapResult, nil
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
