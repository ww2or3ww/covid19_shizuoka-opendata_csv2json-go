package csv2json

/*
陽性患者数

csv
No,全国地方公共団体コード,都道府県名,市区町村名,公表_年月日,曜日,発症_年月日,患者_居住地,患者_年代,患者_性別,患者_職業,患者_状態,患者_症状,患者_渡航歴の有無フラグ,退院済フラグ,備考
1,221309,静岡県,浜松市,2020-03-28,土,,浜北区,,男性,自営業,軽症,,0,1,
2,221309,静岡県,浜松市,2020-04-01,水,2020-03-24,中区,30歳代,男性,会社員,軽症,,0,1,

json
  "patients_summary": {
    "date": "2021/06/12 15:01",
    "data": [
      {
        "日付": "2020-01-29T08:00:00.000Z",
        "小計": 0
      },
      {
        "日付": "2020-01-30T08:00:00.000Z",
        "小計": 0
      },
      :
      :
      {
      	"日付": "2021-06-12T08:00:00.000Z",
        "小計": 3
      },
      {
        "日付": "2021-06-13T08:00:00.000Z",
        "小計": 0
      }
    ]
  },
*/

import (
	"errors"
	"time"

	"github.com/go-gota/gota/dataframe"
)

const keyPatientsSummaryDateOfPublicate = "公表_年月日"

type (
	PatientSummaryData struct {
		Date     string `json:"日付"`
		Subtotal int    `json:"小計"`
	}
	PatientsSummary struct {
		Date string               `json:"date"`
		Data []PatientSummaryData `json:"data"`
	}
)

func patientsSummary(df *dataframe.DataFrame, dtUpdated time.Time, dtEnd time.Time) (*PatientsSummary, error) {
	dfSelected := df.Select(keyPatientsSummaryDateOfPublicate)
	if df.Err != nil {
		return nil, df.Err
	}

	// 日付ごとにカウントアップ
	maps := make(map[string]int)
	for _, v := range dfSelected.Maps() {
		dateOfPublicate, ok := v[keyPatientsSummaryDateOfPublicate].(string)
		if !ok {
			return nil, errors.New("unable to cast patients summary date of publicate to string")
		}
		num := maps[dateOfPublicate]
		num++
		maps[dateOfPublicate] = num
	}

	// 2020-01-29 から 指定日までの 日ごとの配列を作成
	dtStart, _ := time.Parse("2006-01-02", "2020-01-29")
	diffDate := dtEnd.Sub(dtStart)
	days := int(diffDate.Hours()) / 24
	var dataList = make([]PatientSummaryData, days+1)

	// 2020-01-29 から 指定日までの 日ごとデータを作成して配列にセット
	i := 0
	for d := dtStart; d.Unix() < dtEnd.Unix(); d = d.AddDate(0, 0, 1) {
		keyDate := d.Format("2006-01-02")
		var data PatientSummaryData
		data.Date = keyDate + "T08:00:00.000Z"
		data.Subtotal = maps[keyDate]
		dataList[i] = data
		i++
	}

	ps := &PatientsSummary{
		Date: dtUpdated.Format("2006/01/02 15:04"),
		Data: dataList,
	}

	return ps, nil
}
