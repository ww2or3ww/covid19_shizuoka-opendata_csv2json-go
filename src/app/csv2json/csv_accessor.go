package csv2json

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/go-gota/gota/dataframe"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"app/utils/apiutil"
	"app/utils/logger"
)

type csvAccessor struct{}

func NewCsvAccessor() *csvAccessor {
	return &csvAccessor{}
}

// GetCSVDataFrameFromApi はAPIコールによりCSVデータを取得する
func (ca *csvAccessor) GetCSVDataFrameFromApi(apiAddress string) (*dataframe.DataFrame, time.Time, error) {

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
	prCsv := transform.NewReader(ioReaderCsv, japanese.ShiftJIS.NewDecoder())
	dfCsv := dataframe.ReadCSV(prCsv, dataframe.WithDelimiter(','), dataframe.HasHeader(true))

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

// GetTimeNow は今日の日付を取得する
func (ca *csvAccessor) GetTimeNow() time.Time {
	return time.Now()
}
