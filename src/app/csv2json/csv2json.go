package csv2json

import (
	"app/utils/apiutil"
	"app/utils/logger"
	"app/utils/maputil"
	"encoding/json"
	"fmt"
	"github.com/Songmu/go-httpdate"
	"github.com/go-gota/gota/dataframe"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"path/filepath"
	"strings"
	"time"
)

func Process(apiAddress string, queryStrPrm string) *map[string]interface{} {
	var mapResult map[string]interface{}
	logger.Infos(apiAddress, queryStrPrm)

	dtLastUpdate := time.Date(2000, 1, 1, 1, 1, 0, 0, time.Local)
	hasError := false
	types := strings.Split(queryStrPrm, ",")
	for index, value := range types {
		values := strings.Split(value, ":")
		key := values[0]
		apiId := values[1]
		logger.Infof("%d, key=%s, id=%s", index, key, apiId)
		dfCsv, dtUpdated, err := getCSVDataFrameFromApi(fmt.Sprintf("%s?id=%s", apiAddress, apiId))
		if err != nil {
			logger.Errors(err)
			hasError = true
		}
		var mapTmp *map[string]interface{}
		switch key {
		case "main_summary":
			if mapResult[key] == nil {
				mapTmp = mainSummary(dfCsv, dtUpdated)
			} else {
				mapMainSummary := mapResult[key].(map[string]interface{})
				mainSummaryTry2Merge4xx(dfCsv, &mapMainSummary)
				continue
			}
		default:
			mapTmp = mapNotSupported(key)
		}
		logger.Infos(mapTmp)
		if mapTmp != nil {
			mapWithKey := map[string]interface{}{key: *mapTmp}
			mapResult = maputil.MergeMaps(mapResult, mapWithKey)
			if dtUpdated.After(dtLastUpdate) {
				dtLastUpdate = dtUpdated
			}
		}
	}
	mapResult["hasError"] = hasError
	mapResult["lastUpdate"] = dtLastUpdate

	return &mapResult
}

func mapNotSupported(key string) *map[string]interface{} {
	jsonStr := fmt.Sprintf(`
	  {
	    "%s": "not supported..."
	  }
	`, key)
	var mapResult = make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &mapResult)
	if err != nil {
		logger.Errors(err)
	}
	return &mapResult
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

	valueResult, _ := (*mapBody)["result"]
	mapResult := valueResult.(map[string]interface{})
	valueResources := mapResult["resources"]
	listResources := valueResources.([]interface{})
	for _, resource := range listResources {
		mapResource := resource.(map[string]interface{})
		downloadUrl, _ := mapResource["download_url"]
		ext := strings.ToLower(filepath.Ext(downloadUrl.(string)))
		if ext == ".csv" {
			csvAddress = downloadUrl.(string)
			updated, _ := mapResource["updated"]
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
