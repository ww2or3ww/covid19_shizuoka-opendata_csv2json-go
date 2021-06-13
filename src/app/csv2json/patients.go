package csv2json

/*
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
	"app/utils/logger"
	"app/utils/maputil"
	"encoding/json"
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"github.com/guregu/null"
	"time"
)

const keyDay = "公表_年月日"
const keyCity = "市区町村名"
const keyResidence = "患者_居住地"
const keyAge = "患者_年代"
const keySex = "患者_性別"
const keyDischarge = "退院済フラグ"

type (
	Patient struct {
		Release   string      `json:"リリース日"`
		Residence string      `json:"居住地"`
		Age       string      `json:"年代"`
		Sex       string      `json:"性別"`
		Discharge null.String `json:"退院"`
		Date      string      `json:"date"`
	}
)

func patients(df *dataframe.DataFrame, dtUpdated time.Time) *map[string]interface{} {
	dfSelected := df.Select([]string{keyDay, keyCity, keyResidence, keyAge, keySex, keyDischarge})

	logger.Infos(len(dfSelected.Maps()))
	var maps = make([]map[string]interface{}, len(dfSelected.Maps()), len(dfSelected.Maps()))
	for i, v := range dfSelected.Maps() {
		residence := v[keyResidence]
		if residence == "" {
			residence = "--"
		}
		age := v[keyAge]
		if age == "" {
			age = "不明"
		}
		sex := v[keySex]
		if sex == "" {
			sex = "不明"
		}
		var discharge null.String
		if v[keyDischarge] == 1 {
			discharge = null.NewString(`"○"`, true)
		}

		var patientData Patient
		patientData.Release = fmt.Sprintf(`%s`, v[keyDay])
		patientData.Residence = fmt.Sprintf(`%s %s`, v[keyCity], residence)
		patientData.Age = fmt.Sprintf(`%s`, age)
		patientData.Sex = fmt.Sprintf(`%s`, sex)
		patientData.Discharge = discharge
		patientData.Date = fmt.Sprintf(`%s`, v[keyDay])

		btData, err := json.Marshal(patientData)
		jsonStr := string(btData)

		var mapTmp map[string]interface{}
		err = json.Unmarshal([]byte(jsonStr), &mapTmp)
		if err != nil {
			logger.Errors(err)
		}

		maps[i] = mapTmp
	}
	mapsData := make(map[string]interface{}, 0)
	mapsData["data"] = maps

	jsonStr := fmt.Sprintf(`
	  {
      "date": "%s"
	  }
	`, dtUpdated.Format("2006/01/02 15:04"))
	var mapResult = make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &mapResult)
	if err != nil {
		logger.Errors(err)
	}

	mapResult = maputil.MergeMaps(mapsData, mapResult)

	return &mapResult
}
