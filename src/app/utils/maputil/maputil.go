package maputil

func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	mapMerged := make(map[string]interface{}, 0)
	for _, m := range maps {
		for key, value := range m {
			mapMerged[key] = value
		}
	}
	return mapMerged
}
