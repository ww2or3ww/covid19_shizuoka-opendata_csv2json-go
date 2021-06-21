package csv2json

/*
csv
No,全国地方公共団体コード,都道府県名,市区町村名,公表_年月日,曜日,発症_年月日,患者_居住地,患者_年代,患者_性別,患者_職業,患者_状態,患者_症状,患者_渡航歴の有無フラグ,退院済フラグ,備考
1,221309,静岡県,浜松市,2020-03-28,土,,浜北区,,男性,自営業,軽症,,0,1,
2,221309,静岡県,浜松市,2020-04-01,水,2020-03-24,中区,30歳代,男性,会社員,軽症,,0,1,

json
  "main_summary": {
    "date": "2021/06/12 15:01",
    "children": [
      {
        "attr": "陽性患者数",
        "value": 2180,
        "children": [
          {
            "attr": "入院中",
            "value": 215,
            "children": [
              {
                "attr": "軽症・中等症",
                "value": 214
              },
              {
                "attr": "重症",
                "value": 1
              }
            ]
          },
          {
            "attr": "退院",
            "value": 1916
          },
          {
            "attr": "死亡",
            "value": 49
          }
        ]
      }
    ]
  },
*/

import (
	"errors"
	"time"

	"github.com/go-gota/gota/dataframe"
)

const keyMainSummaryPatientStatus = "患者_状態"
const keyMainSummaryDischaFlg = "退院済フラグ"
const keyMainSummaryNumberOfDeath = "死亡者人数"

type MainSummaryData struct {
	Attr     string            `json:"attr"`
	Value    int               `json:"value"`
	Children []MainSummaryData `json:"children,omitempty"`
}

type MainSummary struct {
	Date     string            `json:"date"`
	Children []MainSummaryData `json:"children"`
}

func isMildStatus(patientStatus string) bool {
	return (patientStatus == "軽症" || patientStatus == "中等症" || patientStatus == "無症状")
}

func isServStatus(patientStatus string) bool {
	return (patientStatus == "重症")
}

// [検査陽性患者の属性csv]から[検査陽性者の状況]を生成する
func mainSummary(df *dataframe.DataFrame, dtUpdated time.Time) (*MainSummary, error) {
	var sumPosi = 0   // 陽性患者数
	var sumHosp = 0   // 入院中
	var sumMild = 0   // 軽症・中軽症 + 無症状
	var sumServ = 0   // 重症
	var sumDischa = 0 // 退院

	dfSelected := df.Select([]string{keyMainSummaryPatientStatus, keyMainSummaryDischaFlg})
	if df.Err != nil {
		return nil, df.Err
	}
	for _, v := range dfSelected.Maps() {
		patientStatus, ok := v[keyMainSummaryPatientStatus].(string)
		if !ok {
			return nil, errors.New("unable to cast main summary patient status to string")
		}
		dischanFlg := v[keyMainSummaryDischaFlg]

		// 陽性患者数
		sumPosi += 1

		// 入院患者数
		if dischanFlg == 0 {
			sumHosp += 1
		}

		if isMildStatus(patientStatus) && dischanFlg == 0 {
			// 軽症・中等症者数
			sumMild += 1
		} else if isServStatus(patientStatus) && dischanFlg == 0 {
			// 重傷者数
			sumServ += 1
		} else if dischanFlg == 1 {
			// 退院者数
			sumDischa += 1
		}
	}

	m := &MainSummary{
		Date: dtUpdated.Format("2006/01/02 15:04"),
		Children: []MainSummaryData{
			{
				Attr:  "陽性患者数",
				Value: sumPosi,
				Children: []MainSummaryData{
					{
						Attr:  "入院中",
						Value: sumHosp,
						Children: []MainSummaryData{
							{
								Attr:  "軽症・中等症",
								Value: sumMild,
							},
							{
								Attr:  "重症",
								Value: sumServ,
							},
						},
					},
					{
						Attr:  "退院",
						Value: sumDischa,
					},
					{
						Attr:  "死亡",
						Value: 0,
					},
				},
			},
		},
	}

	return m, nil
}

// [検査陽性者の状況] の 死亡者数 を [陽性患者数csv] からカウントして取得する。
// ([検査陽性患者の属性csv]だけで死亡者を表現すると、死亡者の特定に繋がってしまうため)
// また、退院数から死亡者数を減算する。
func mainSummaryTry2Merge4Deth(df *dataframe.DataFrame, mainSummary *MainSummary) error {
	var numberOfDeth = 0 // 死亡者数
	dfSelected := df.Select(keyMainSummaryNumberOfDeath)
	for _, v := range dfSelected.Maps() {
		n, ok := v[keyMainSummaryNumberOfDeath].(int)
		if !ok {
			return errors.New("unable to cast main summery number of death")
		}
		numberOfDeth = numberOfDeth + n
	}

	if len(mainSummary.Children) != 1 || len(mainSummary.Children[0].Children) != 3 {
		return errors.New("unexpected main summary structure")
	}

	// 死亡の値を設定
	mainSummary.Children[0].Children[2].Value = numberOfDeth

	// 退院の値から死亡の値を減算
	mainSummary.Children[0].Children[1].Value -= numberOfDeth

	return nil
}
