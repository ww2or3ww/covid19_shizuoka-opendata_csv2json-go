package csv2json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/go-gota/gota/dataframe"

	"app/utils/logger"
)

const defApiAddress = "https://opendata.pref.shizuoka.jp/api/package_show"

func init() {
	logLv := logger.Error
	envLogLv := os.Getenv("LOG_LEVEL")
	logger.Infos(envLogLv)
	if envLogLv != "" {
		n, _ := strconv.Atoi(envLogLv)
		logLv = logger.LogLv(n)
	}
	logger.LogInitialize(logLv, 25)
}

func TestProcess(t *testing.T) {
	isUseMock := true

	type args struct {
		apiAddress  string
		queryStrPrm string
	}
	type result struct {
		hasError bool
	}
	tests := []struct {
		name    string
		args    args
		useMock bool
		result  result
	}{
		{
			name: "normal",
			args: args{
				apiAddress:  defApiAddress,
				queryStrPrm: "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,main_summary:92f9ebcd-a3f1-4d5d-899b-d69214294a45,patients:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,patients_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,inspection_persons:d4827176-d887-412a-9344-f84f161786a2,contacts:1b57f2c0-081e-4664-ba28-9cce56d0b314",
				//queryStrPrm: "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,inspection_persons:d4827176-d887-412a-9344-f84f161786a2",
			},
			useMock: isUseMock,
			result: result{
				hasError: false,
			},
		},
		{
			name: "hasError : invalid qyery param",
			args: args{
				apiAddress:  defApiAddress,
				queryStrPrm: "main_summary",
			},
			useMock: isUseMock,
			result: result{
				hasError: false,
			},
		},
		{
			name: "hasError : invalid address",
			args: args{
				apiAddress:  "xxx",
				queryStrPrm: "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1",
			},
			useMock: isUseMock,
			result: result{
				hasError: false,
			},
		},
	}

	for _, tt := range tests {
		var c2j *Csv2Json
		if tt.useMock {
			c2j = NewCsv2Json(newCsvAccessorTest())
		} else {
			c2j = NewCsv2Json(NewCsvAccessor())
		}
		t.Run(tt.name, func(t *testing.T) {
			mapResult, err := c2j.Process(tt.args.apiAddress, tt.args.queryStrPrm)
			if err != nil {
				if !tt.result.hasError {
					t.Errorf(tt.name)
				}
				return
			}
			jsonIndent, _ := json.MarshalIndent(mapResult, "", "   ")
			logger.Debugs(string(jsonIndent))

			if mapResult == nil {
				t.Errorf(tt.name)
			}
		})
	}
}

type csvAccessorTest struct{}

func newCsvAccessorTest() accessor {
	return &csvAccessorTest{}
}

func (c *csvAccessorTest) GetTimeNow() time.Time {
	// テストデータの最新日は 2020/4/8 としている
	dtNow, _ := time.Parse("2006-01-02 15:04", "2020-04-08 12:00")
	return dtNow
}

func (c *csvAccessorTest) GetCSVDataFrameFromApi(apiAddress string) (*dataframe.DataFrame, time.Time, error) {
	var dfCsv *dataframe.DataFrame
	var updatedDateTime time.Time
	var err error

	switch apiAddress {
	case "https://opendata.pref.shizuoka.jp/api/package_show?id=5ab47071-3651-457c-ae2b-bfb8fdbe1af1":
		dfCsv = getDataFrameFromCsvFile("./testdata/patients.csv")
	case "https://opendata.pref.shizuoka.jp/api/package_show?id=92f9ebcd-a3f1-4d5d-899b-d69214294a45":
		dfCsv = getDataFrameFromCsvFile("./testdata/patients_summary.csv")
	case "https://opendata.pref.shizuoka.jp/api/package_show?id=d4827176-d887-412a-9344-f84f161786a2":
		dfCsv = getDataFrameFromCsvFile("./testdata/test_people.csv")
	case "https://opendata.pref.shizuoka.jp/api/package_show?id=1b57f2c0-081e-4664-ba28-9cce56d0b314":
		dfCsv = getDataFrameFromCsvFile("./testdata/call_center.csv")
	default:
		logger.Errorf("not supported address : %s", apiAddress)
		dfCsv = nil
		err = fmt.Errorf("unknown api address : %v", apiAddress)
	}

	strDateTime := "2021/06/15 15:18"
	updatedDateTime, _ = httpdate.Str2Time(strDateTime, nil)

	return dfCsv, updatedDateTime, err
}

func getDataFrameFromCsvFile(fileName string) *dataframe.DataFrame {
	content, _ := ioutil.ReadFile(fileName)
	ioContent := strings.NewReader(string(content))

	df := dataframe.ReadCSV(ioContent)
	return &df
}
