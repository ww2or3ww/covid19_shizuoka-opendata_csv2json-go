package csv2json

import (
	"app/utils/apiutil"
	"app/utils/logger"
	"app/utils/maputil"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/go-gota/gota/dataframe"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type (
	CsvData struct {
		DfCsv     *dataframe.DataFrame
		DtUpdated time.Time
	}
)

// 同じCSVデータを何度も読みにいかないためにバックアップしておくための変数
// key	: csv address
var mapCSVDataBackup = make(map[string](*CsvData))

// オープンデータのCSVをJSONに変換する処理
func Process(apiAddress string, queryStrPrm string) *map[string]interface{} {
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
		csvData, err := getCSVDataFrame(fmt.Sprintf("%s?id=%s", apiAddress, apiId))

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
				mapTmp = patientsSummary(csvData.DfCsv, csvData.DtUpdated)
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

func getCSVDataFrame(apiAddress string) (*CsvData, error) {
	data := mapCSVDataBackup[apiAddress]
	var err error
	if data == nil {
		data = &CsvData{}
		data.DfCsv, data.DtUpdated, err = getCSVDataFrameFromApi(apiAddress)
		mapCSVDataBackup[apiAddress] = data
	}
	return data, err
}

func getCSVDataFrameFromApi(apiAddress string) (*dataframe.DataFrame, time.Time, error) {

	// get json from api
	mapBody, err := apiutil.GetJsonMapFromResponseBody(apiAddress)
	if err != nil {
		logger.Errors(err)
		return nil, time.Time{}, err
	}

	// get csv address from json
	csvAddress, updatedDateTime, err := getCsvAddressFromBody(mapBody)
	if err != nil {
		logger.Errors(err)
		return nil, time.Time{}, err
	}
	logger.Infof("csv address = %v", csvAddress)
	logger.Infof("update time = %v", updatedDateTime)

	// get bytes data from csv
	bytesCsv, err := apiutil.GetBytesFromResponseBody(csvAddress)
	if err != nil {
		logger.Errors(err)
		return nil, time.Time{}, err
	}

	// convert to dataframe from csv bytes data
	ioReaderCsv := strings.NewReader(string(bytesCsv))
	strCsv := transform.NewReader(ioReaderCsv, japanese.ShiftJIS.NewDecoder())
	dfCsv := dataframe.ReadCSV(strCsv, dataframe.WithDelimiter(','), dataframe.HasHeader(true))

	return &dfCsv, updatedDateTime, nil
}

// mapBody["result"]["resources"][n]["download_url"](*.csv)
func getCsvAddressFromBody(mapBody *map[string]interface{}) (csvAddress string, updatedDateTime time.Time, errOut error) {
	csvAddress = ""
	errOut = nil

	mapResult := ((*mapBody)["result"]).(map[string]interface{})
	listResources := (mapResult["resources"]).([]interface{})
	for _, resource := range listResources {
		mapResource := resource.(map[string]interface{})
		downloadUrl := mapResource["download_url"]
		ext := strings.ToLower(filepath.Ext(downloadUrl.(string)))
		if ext == ".csv" {
			csvAddress = downloadUrl.(string)
			updated := mapResource["updated"]
			updatedDateTime, _ = httpdate.Str2Time(updated.(string), nil)
			break
		}
	}
	if csvAddress == "" {
		errMsg := "not found .csv resource from body"
		logger.Errors(errMsg)
		return "", updatedDateTime, fmt.Errorf("%s", errMsg)
	}

	return csvAddress, updatedDateTime, errOut
}
