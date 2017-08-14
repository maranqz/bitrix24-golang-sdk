package bitrix24



type ResponseBitrix24 map[string]interface{}

/*type ResponseBitrix24 struct {
	response struct {
		result       map[string]string
		result_error map[string]string
		result_total map[string]string
		result_next  map[string]string
	}
}*/

type BatchResponseBitrix24 struct {
	result struct {
		result       []map[string]string
		result_error []map[string]string
		result_total []map[string]string
		result_next  []map[string]string
	}
}
