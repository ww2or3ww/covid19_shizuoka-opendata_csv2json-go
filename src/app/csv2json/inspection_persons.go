package csv2json

/*
検査実施人数

csv
実施_年月日,全国地方公共団体コード,都道府県名,市区町村名,検査実施_人数,備考
2020-01-29,221309,静岡県,浜松市,0,
2020-01-30,221309,静岡県,浜松市,0,

json
  "inspection_persons": {
    "date": "2021/06/13 18:06",
    "labels": [
      "2020-01-29T08:00:00.000Z",
      "2020-01-30T08:00:00.000Z",
      :
      :
      "2021-06-11T08:00:00.000Z",
      "2021-06-12T08:00:00.000Z"
    ],
    "datasets": [
      {
        "label": "PCR検査実施人数",
        "data": [
          0,
          0,
      :
      :
          79,
          17
        ]
      }
    ]
  },
*/

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-gota/gota/dataframe"
)

const keyInspectPersonsDate = "実施_年月日"
const keyInspectPersonsNumOfPeople = "検査実施_人数"

type (
	InspectionDataset struct {
		Label string `json:"label"`
		Data  []int  `json:"data"`
	}
	InspectionPersons struct {
		Date     string              `json:"date"`
		Labels   []string            `json:"labels"`
		Datasets []InspectionDataset `json:"datasets"`
	}
)

func inspectionPersons(df *dataframe.DataFrame, dtUpdated time.Time) (*InspectionPersons, error) {
	dfSelected := df.Select([]string{keyInspectPersonsDate, keyInspectPersonsNumOfPeople})
	if df.Err != nil {
		return nil, df.Err
	}

	ip := &InspectionPersons{
		Date:   dtUpdated.Format("2006/01/02 15:04"),
		Labels: make([]string, len(dfSelected.Maps())),
		Datasets: []InspectionDataset{
			{
				Label: "PCR検査実施人数",
				Data:  make([]int, len(dfSelected.Maps())),
			},
		},
	}

	// 行ごとのデータを取得して配列へセット
	for i, v := range dfSelected.Maps() {
		ip.Labels[i] = fmt.Sprintf("%s%s", v[keyInspectPersonsDate], "T08:00:00.000Z")
		n, ok := v[keyInspectPersonsNumOfPeople].(int)
		if !ok {
			return nil, errors.New("unable to cast inspect persons num of people to int")
		}
		ip.Datasets[0].Data[i] = n
	}

	return ip, nil
}
