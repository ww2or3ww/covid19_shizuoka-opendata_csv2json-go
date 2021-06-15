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
	"app/utils/maputil"
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
	Inspectionpersons struct {
		Date     string              `json:"date"`
		Labels   []string            `json:"labels"`
		Datasets []InspectionDataset `json:"datasets"`
	}
)

func inspectionPersons(df *dataframe.DataFrame, dtUpdated time.Time) *map[string]interface{} {
	dfSelected := df.Select([]string{keyInspectPersonsDate, keyInspectPersonsNumOfPeople})

	// 行ごとのデータを取得して配列へセット
	dateList := make([]string, len(dfSelected.Maps()))
	numList := make([]int, len(dfSelected.Maps()))
	for i, v := range dfSelected.Maps() {
		dateList[i] = fmt.Sprintf("%s%s", v[keyInspectPersonsDate], "T08:00:00.000Z")
		numList[i] = v[keyInspectPersonsNumOfPeople].(int)
	}

	datasetList := make([]InspectionDataset, 1)
	datasetList[0].Data = numList
	datasetList[0].Label = "PCR検査実施人数"

	var stResult Inspectionpersons
	stResult.Date = dtUpdated.Format("2006/01/02 15:04")
	stResult.Labels = dateList
	stResult.Datasets = datasetList

	return maputil.StructToMap(stResult)
}
