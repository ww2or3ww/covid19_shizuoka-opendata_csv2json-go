package csv2json

import (
	"app/utils/logger"
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

type Result struct {
	Contacts          *Contacts          `json:"contacts"`
	InspectionPersons *InspectionPersons `json:"inspection_persons"`
	MainSummary       *MainSummary       `json:"main_summary"`
	Patients          *Patients          `json:"patients"`
	PatientsSummary   *PatientsSummary   `json:"patients_summary"`
	Value             int                `json:"value"`
	HasError          bool               `json:"hasError"`
	LastUpdate        string             `json:"lastUpdate"`
}

type Csv2Json interface {
	Process(apiAddress string, queryStrPrm string) (*Result, error)
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
func (c2j *csv2Json) Process(apiAddress string, queryStrPrm string) (*Result, error) {
	r := &Result{
		Value:    0,
		HasError: false,
	}
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
			if r.MainSummary == nil {
				r.MainSummary, err = mainSummary(csvData.DfCsv, csvData.DtUpdated)
				if err != nil {
					return nil, err
				}
			} else {
				err = mainSummaryTry2Merge4Deth(csvData.DfCsv, r.MainSummary)
				if err != nil {
					return nil, err
				}
			}
		case "patients":
			r.Patients, err = patients(csvData.DfCsv, csvData.DtUpdated)
			if err != nil {
				return nil, err
			}
		case "patients_summary":
			r.PatientsSummary, err = patientsSummary(csvData.DfCsv, csvData.DtUpdated, c2j.csvAccessor.GetTimeNow())
			if err != nil {
				return nil, err
			}
		case "inspection_persons":
			r.InspectionPersons, err = inspectionPersons(csvData.DfCsv, csvData.DtUpdated)
			if err != nil {
				return nil, err
			}
		case "contacts":
			r.Contacts, err = contacts(csvData.DfCsv, csvData.DtUpdated)
			if err != nil {
				return nil, err
			}
		default:
			message := "not supported..."
			logger.Errors(key, message)
			return nil, errors.New("not supported")
		}

		if csvData.DtUpdated.After(dtLastUpdate) {
			dtLastUpdate = csvData.DtUpdated
		}

		logger.Infof("%s time = %d milliseconds", value, time.Since(timeStart).Milliseconds())
	}

	r.LastUpdate = dtLastUpdate.Format("2006/01/02 15:04")

	return r, nil
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
