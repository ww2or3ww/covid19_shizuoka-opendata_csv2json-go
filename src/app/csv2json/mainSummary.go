package csv2json

import (
	"app/utils/logger"
	"encoding/json"
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"time"
)

const keyPatientStatus = "患者_状態"
const keyDischaFlg = "退院済フラグ"
const keyNumberOfXX = "死亡者人数"

func isMildStatus(patientStatus string) bool {
	return (patientStatus == "軽症" || patientStatus == "中等症" || patientStatus == "無症状")
}

func isServStatus(patientStatus string) bool {
	return (patientStatus == "重症")
}

func mainSummary(df *dataframe.DataFrame, dtUpdated time.Time) *map[string]interface{} {
	var sumPosi = 0   // 陽性患者数
	var sumHosp = 0   // 入院中
	var sumMild = 0   // 軽症・中軽症 + 無症状
	var sumServ = 0   // 重症
	var sumDischa = 0 // 退院

	dfSelected := df.Select([]string{keyPatientStatus, keyDischaFlg})
	for _, v := range dfSelected.Maps() {
		patientStatus := v[keyPatientStatus]
		dischanFlg := v[keyDischaFlg]

		// 陽性患者数
		sumPosi += 1

		// 入院患者数
		if dischanFlg == 0 {
			sumHosp += 1
		}

		if isMildStatus(patientStatus.(string)) && dischanFlg == 0 {
			// 軽症・中等症者数
			sumMild += 1
		} else if isServStatus(patientStatus.(string)) && dischanFlg == 0 {
			// 重傷者数
			sumServ += 1
		} else if dischanFlg == 1 {
			// 退院者数
			sumDischa += 1
		}
	}

	jsonStr := fmt.Sprintf(`
	  {
	    "date": "2021/06/07 18:56",
	    "children": [
	      {
	        "attr": "陽性患者数",
	        "value": %d,
	        "children": [
	          {
	            "attr": "入院中",
	            "value": %d,
	            "children": [
	              {
	                "attr": "軽症・中等症",
	                "value": %d
	              },
	              {
	                "attr": "重症",
	                "value": %d
	              }
	            ]
	          },
	          {
	            "attr": "退院",
	            "value": %d
	          },
	          {
	            "attr": "死亡",
	            "value": 0
	          }
	        ]
	      }
	    ]
	  }
	`, sumPosi, sumHosp, sumMild, sumServ, sumDischa)

	var mapResult = make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &mapResult)
	if err != nil {
		logger.Errors(err)
	}

	return &mapResult
}

func mainSummaryTry2Merge4xx(df *dataframe.DataFrame, mapMainSummary *map[string]interface{}) {
	var numberOfDeth = 0 // 死亡者数
	dfSelected := df.Select(keyNumberOfXX)
	for _, v := range dfSelected.Maps() {
		numberOfDeth = numberOfDeth + v[keyNumberOfXX].(int)
	}

	aryChildren1 := (*mapMainSummary)["children"].([]interface{})
	aryChildren2 := aryChildren1[0].(map[string]interface{})["children"].([]interface{})

	// 死亡の値を設定
	mapDeath := aryChildren2[2].(map[string]interface{})
	mapDeath["value"] = numberOfDeth

	// 退院の値から死亡の値を減算
	mapDischa := aryChildren2[1].(map[string]interface{})
	mapDischa["value"] = mapDischa["value"].(float64) - float64(numberOfDeth)
}
