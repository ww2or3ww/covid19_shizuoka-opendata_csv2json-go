package csv2json

/*
新型コロナウイルス感染症に関する相談件数

csv
全国地方公共団体コード,都道府県名,市区町村名,受付_年月日,曜日,相談件数
221309,静岡県,浜松市,2020-01-29,水,27
221309,静岡県,浜松市,2020-01-30,木,52

json
  "contacts": {
    "date": "2021/06/11 15:45",
    "data": [
      {
        "日付": "2020-01-29T08:00:00.000Z",
        "小計": 27
      },
      {
        "日付": "2020-01-30T08:00:00.000Z",
        "小計": 52
      },
      :
      :
      {
        "日付": "2021-06-09T08:00:00.000Z",
        "小計": 157
      },
      {
        "日付": "2021-06-10T08:00:00.000Z",
        "小計": 128
      }
    ]
  },
*/

import (
	"app/utils/logger"
	"app/utils/maputil"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-gota/gota/dataframe"
)

const keyContactsDateOfReceipt = "受付_年月日"
const keyContactsNumOfConsulted = "相談件数"

type (
	Contacts struct {
		Date     string `json:"日付"`
		Subtotal int    `json:"小計"`
	}
)

func contacts(df *dataframe.DataFrame, dtUpdated time.Time) *map[string]interface{} {
	dfSelected := df.Select([]string{keyContactsDateOfReceipt, keyContactsNumOfConsulted})

	// 日付ごとにカウントアップ
	maps := make(map[string]int)
	for _, v := range dfSelected.Maps() {
		dateOfReceipt := v[keyContactsDateOfReceipt]
		numOfConsulted := v[keyContactsNumOfConsulted]
		maps[dateOfReceipt.(string)] = numOfConsulted.(int)
	}

	// 2020-01-29 から 今日までの 日ごとの配列を作成
	startDate, _ := time.Parse("2006-01-02", "2020-01-29")
	today := time.Now()
	diffDate := today.Sub(startDate)
	days := int(diffDate.Hours())/24 + 1
	var dataList = make([]Contacts, days)

	// 2020-01-29 から 今日までの 日ごとデータを作成して配列にセット
	i := 0
	for d := startDate; d.Unix() < time.Now().Unix(); d = d.AddDate(0, 0, 1) {
		keyDate := d.Format("2006-01-02")
		var data Contacts
		data.Date = keyDate + "T08:00:00.000Z"
		data.Subtotal = maps[keyDate]
		dataList[i] = data
		i++
	}

	// data
	mapsData := make(map[string]interface{})
	mapsData["data"] = dataList

	// date
	jsonStr := fmt.Sprintf(`
	  {
      "date": "%s"
	  }
	`, dtUpdated.Format("2006/01/02 15:04"))
	var mapDate = make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &mapDate)
	if err != nil {
		logger.Errors(err)
	}

	// data と date を マージ
	mapResult := maputil.MergeMaps(mapsData, mapDate)

	return &mapResult
}
