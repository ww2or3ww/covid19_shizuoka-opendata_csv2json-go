package csv2json

import (
	"app/utils/logger"
	"encoding/json"
	"fmt"
	"github.com/Songmu/go-httpdate"
	"github.com/go-gota/gota/dataframe"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
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
		logger.Infos("%d, key=%s, id=%s", index, key, apiId)
		dfCsv, dtUpdated, err := getCSVData(fmt.Sprintf("%s?id=%s", apiAddress, apiId))
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
			mapResult = mergeMaps(mapResult, mapWithKey)
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

func mergeMaps(m ...map[string]interface{}) map[string]interface{} {
	ans := make(map[string]interface{}, 0)
	for _, c := range m {
		for k, v := range c {
			ans[k] = v
		}
	}
	return ans
}

func getCSVData(apiAddress string) (*dataframe.DataFrame, time.Time, error) {
	mapBody, err := getJsonMapFromResponseBody(apiAddress)
	if err != nil {
		logger.Errors(err)
		return nil, time.Time{}, err
	}

	csvAddress, updatedDateTime, err := getCsvAddressFromBody(mapBody)
	if err != nil {
		logger.Errors(err)
		return nil, time.Time{}, err
	}
	logger.Infos(csvAddress)
	logger.Infos(updatedDateTime)

	bytesCsv, err := getBytesFromResponseBody(csvAddress)
	if err != nil {
		logger.Errors(err)
		return nil, time.Time{}, err
	}
	ioReaderCsv := strings.NewReader(string(bytesCsv))
	strCsv := transform.NewReader(ioReaderCsv, japanese.ShiftJIS.NewDecoder())
	dfCsv := dataframe.ReadCSV(strCsv, dataframe.WithDelimiter(','), dataframe.HasHeader(true))

	return &dfCsv, updatedDateTime, nil
}

func getJsonMapFromResponseBody(apiAddress string) (*map[string]interface{}, error) {

	bytesBody, err := getBytesFromResponseBody(apiAddress)
	if err != nil {
		logger.Errors(err)
		return nil, err
	}

	// convert to map from body string
	var mapJsonData map[string]interface{}
	err = json.Unmarshal([]byte(bytesBody), &mapJsonData)
	if err != nil {
		logger.Errors(err)
		return nil, err
	}

	return &mapJsonData, nil
}

func getBytesFromResponseBody(address string) ([]byte, error) {
	// creeate get request
	request, err := http.NewRequest("GET", address, nil)
	if err != nil {
		logger.Errors(err)
		return nil, err
	}

	// do get response
	client := new(http.Client)
	response, err := client.Do(request)
	if err != nil {
		logger.Errors(err)
		return nil, err
	}
	defer response.Body.Close()

	// check status code
	logger.Debugs(response.Status)
	if response.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("response.Status : %v", response.Status)
		logger.Errorf(errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// read body bytes
	bytesBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Errors(err)
	}

	return bytesBody, err
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
		logger.Errorf(errMsg)
		return "", updatedDateTime, fmt.Errorf("%s", errMsg)
	}

	return csvAddress, updatedDateTime, errOut
}
