package csv2json

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-gota/gota/dataframe"

	"app/utils/logger"
)

type (
	CsvData struct {
		DfCsv     *dataframe.DataFrame
		DtUpdated time.Time
	}
)

type Result struct {
	Contacts          *Contacts          `json:"contacts,omitempty"`
	InspectionPersons *InspectionPersons `json:"inspection_persons,omitempty"`
	MainSummary       *MainSummary       `json:"main_summary,omitempty"`
	Patients          *Patients          `json:"patients,omitempty"`
	PatientsSummary   *PatientsSummary   `json:"patients_summary,omitempty"`
	Value             int                `json:"value"`
	HasError          bool               `json:"hasError"`
	LastUpdate        string             `json:"lastUpdate"`
}

type accessor interface {
	GetCSVDataFrameFromApi(apiAddress string) (*dataframe.DataFrame, time.Time, error)
	GetTimeNow() time.Time
}

type Csv2Json struct {
	csvAccessor accessor
}

func NewCsv2Json(csvAccessorIn accessor) *Csv2Json {
	return &Csv2Json{csvAccessor: csvAccessorIn}
}

// 同じCSVデータを何度も読みにいかないためにバックアップしておくための変数
// key	: csv address
var mapCSVDataBackup = make(map[string](*CsvData))

// Process はオープンデータのCSVをJSONに変換する
func (c2j *Csv2Json) Process(apiAddress string, queryStrPrm string) (*Result, error) {
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
			r.HasError = true
			logger.Errors(value, "invalid query param...")
			continue
		}

		key := values[0]
		apiId := values[1]
		logger.Infof("%d, key=%s, id=%s", index, key, apiId)
		csvData, err := getCSVDataFrame(fmt.Sprintf("%s?id=%s", apiAddress, apiId), c2j.csvAccessor)

		if err != nil {
			r.HasError = true
			logger.Errors(key, err)
			continue
		}

		switch key {
		case "main_summary":
			if r.MainSummary == nil {
				r.MainSummary, err = mainSummary(csvData.DfCsv, csvData.DtUpdated)
			} else {
				err = mainSummaryTry2Merge4Deth(csvData.DfCsv, r.MainSummary)
			}
		case "patients":
			r.Patients, err = patients(csvData.DfCsv, csvData.DtUpdated)
		case "patients_summary":
			r.PatientsSummary, err = patientsSummary(csvData.DfCsv, csvData.DtUpdated, c2j.csvAccessor.GetTimeNow())
		case "inspection_persons":
			r.InspectionPersons, err = inspectionPersons(csvData.DfCsv, csvData.DtUpdated)
		case "contacts":
			r.Contacts, err = contacts(csvData.DfCsv, csvData.DtUpdated)
		default:
			message := "not supported..."
			logger.Errors(key, message)
			return nil, errors.New("not supported")
		}

		if err != nil {
			r.HasError = true
			logger.Errors(key, err)
			continue
		}

		if csvData.DtUpdated.After(dtLastUpdate) {
			dtLastUpdate = csvData.DtUpdated
		}

		logger.Infof("%s time = %d milliseconds", value, time.Since(timeStart).Milliseconds())
	}

	r.LastUpdate = dtLastUpdate.Format("2006/01/02 15:04")

	return r, nil
}

func getCSVDataFrame(apiAddress string, csvAccessor accessor) (*CsvData, error) {
	data := mapCSVDataBackup[apiAddress]
	var err error
	if data == nil {
		data = &CsvData{}
		data.DfCsv, data.DtUpdated, err = csvAccessor.GetCSVDataFrameFromApi(apiAddress)
		mapCSVDataBackup[apiAddress] = data
	}
	return data, err
}
