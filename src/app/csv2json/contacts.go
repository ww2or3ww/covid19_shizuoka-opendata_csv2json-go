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
	"errors"
	"time"

	"github.com/go-gota/gota/dataframe"
)

const keyContactsDateOfReceipt = "受付_年月日"
const keyContactsNumOfConsulted = "相談件数"

type (
	ContactData struct {
		Date     string `json:"日付"`
		Subtotal int    `json:"小計"`
	}
	Contacts struct {
		Date string        `json:"date"`
		Data []ContactData `json:"data"`
	}
)

func contacts(df *dataframe.DataFrame, dtUpdated time.Time) (*Contacts, error) {
	dfSelected := df.Select([]string{keyContactsDateOfReceipt, keyContactsNumOfConsulted})
	if df.Err != nil {
		return nil, df.Err
	}

	c := &Contacts{
		Date: dtUpdated.Format("2006/01/02 15:04"),
		Data: make([]ContactData, len(dfSelected.Maps())),
	}

	// 日ごとデータを作成して配列にセット
	i := 0
	for _, v := range dfSelected.Maps() {
		dateOfReceipt, ok := v[keyContactsDateOfReceipt].(string)
		if !ok {
			return nil, errors.New("unable to cast contacts date of receipt to string")
		}
		numOfConsulted, ok := v[keyContactsNumOfConsulted].(int)
		if !ok {
			return nil, errors.New("unable to cast contacts num of consulted to int")
		}

		c.Data[i] = ContactData{
			Date:     dateOfReceipt + "T08:00:00.000Z",
			Subtotal: numOfConsulted,
		}
		i++
	}

	return c, nil
}
