package maputil

import "encoding/json"

func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	mapMerged := make(map[string]interface{}, 0)
	for _, m := range maps {
		for key, value := range m {
			mapMerged[key] = value
		}
	}
	return mapMerged
}

func StructToMap(data interface{}) *map[string]interface{} {
	var mapRet map[string]interface{}
	inrec, _ := json.Marshal(data)
	json.Unmarshal(inrec, &mapRet)
	return &mapRet
}
