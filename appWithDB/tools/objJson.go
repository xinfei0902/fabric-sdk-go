package tools

import (
	"encoding/json"
)

func TryParseStringToObj(input []byte) (ret interface{}, err error) {
	if len(input) == 0 {
		ret = string(input)
		return
	}
	switch input[0] {
	case '{':
		obj := make(map[string]interface{})
		err = json.Unmarshal(input, &obj)
		if err != nil {
			break
		}
		ret = obj
	case '[':
		obj := make([]interface{}, 0, 1)
		err = json.Unmarshal(input, &obj)
		if err != nil {
			break
		}
		ret = obj
	default:
		ret = string(input)
	}
	return
}
