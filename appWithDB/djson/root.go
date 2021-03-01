package djson

import (
	jsoniter "github.com/json-iterator/go"
)

var globalCore = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {

}

// Marshal API for json replace
func Marshal(obj interface{}) ([]byte, error) {
	return globalCore.Marshal(obj)
}

// Unmarshal API for json replace
func Unmarshal(datas []byte, obj interface{}) error {
	return globalCore.Unmarshal(datas, obj)
}
