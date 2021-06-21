package csv2json

/*
陽性患者の属性

csv
No,全国地方公共団体コード,都道府県名,市区町村名,公表_年月日,曜日,発症_年月日,患者_居住地,患者_年代,患者_性別,患者_職業,患者_状態,患者_症状,患者_渡航歴の有無フラグ,退院済フラグ,備考
1,221309,静岡県,浜松市,2020-03-28,土,,浜北区,,男性,自営業,軽症,,0,1,
2,221309,静岡県,浜松市,2020-04-01,水,2020-03-24,中区,30歳代,男性,会社員,軽症,,0,1,
3,221309,静岡県,浜松市,2020-04-03,金,2020-04-02,浜北区,,女性,無職,軽症,,0,1,

json
  "patients": {
    "date": "2021/06/12 15:01",
    "data": [
      {
        "リリース日": "2020-03-28T08:00:00.000Z",
        "居住地": "浜松市 浜北区",
        "年代": "不明",
        "性別": "男性",
        "退院": "○",
        "date": "2020-03-28"
      },
      {
        "リリース日": "2020-04-01T08:00:00.000Z",
        "居住地": "浜松市 中区",
        "年代": "30歳代",
        "性別": "男性",
        "退院": "○",
        "date": "2020-04-01"
      },
      :
      :
      {
        "リリース日": "2021-06-12T08:00:00.000Z",
        "居住地": "浜松市 浜北区",
        "年代": "50代",
        "性別": "男性",
        "退院": null,
        "date": "2021-06-12"
      },
      {
        "リリース日": "2021-06-12T08:00:00.000Z",
        "居住地": "浜松市 中区",
        "年代": "60代以上",
        "性別": "男性",
        "退院": null,
        "date": "2021-06-12"
      }
    ]
  },
*/

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/guregu/null"
)

const keyPatientsDay = "公表_年月日"
const keyPatientsCity = "市区町村名"
const keyPatientsResidence = "患者_居住地"
const keyPatientsAge = "患者_年代"
const keyPatientsSex = "患者_性別"
const keyPatientsDischarge = "退院済フラグ"

type (
	PatientData struct {
		Release   string      `json:"リリース日"`
		Residence string      `json:"居住地"`
		Age       string      `json:"年代"`
		Sex       string      `json:"性別"`
		Discharge null.String `json:"退院"`
		Date      string      `json:"date"`
	}
	Patients struct {
		Date string        `json:"date"`
		Data []PatientData `json:"data"`
	}
)

func patients(df *dataframe.DataFrame, dtUpdated time.Time) (*Patients, error) {
	dfSelected := df.Select([]string{keyPatientsDay, keyPatientsCity, keyPatientsResidence, keyPatientsAge, keyPatientsSex, keyPatientsDischarge})
	if df.Err != nil {
		return nil, df.Err
	}

	p := &Patients{
		Date: dtUpdated.Format("2006/01/02 15:04"),
		Data: make([]PatientData, len(dfSelected.Maps())),
	}

	// 行ごとにデータを作成して配列にセット
	for i, v := range dfSelected.Maps() {
		residence := v[keyPatientsResidence]
		if residence == "" {
			residence = "--"
		}
		age := v[keyPatientsAge]
		if age == "" {
			age = "不明"
		}
		sex := v[keyPatientsSex]
		if sex == "" {
			sex = "不明"
		}
		var discharge null.String
		if v[keyPatientsDischarge] == 1 {
			discharge = null.NewString(`○`, true)
		}

		r, ok := v[keyPatientsDay].(string)
		if !ok {
			return nil, errors.New("unable to cast patients day to string")
		}
		p.Data[i] = PatientData{
			Release:   r + "T08:00:00.000Z",
			Residence: fmt.Sprintf(`%s %s`, v[keyPatientsCity], residence),
			Age:       fmt.Sprintf(`%s`, age),
			Sex:       fmt.Sprintf(`%s`, sex),
			Discharge: discharge,
			Date:      fmt.Sprintf(`%s`, v[keyPatientsDay]),
		}
	}

	return p, nil
}
