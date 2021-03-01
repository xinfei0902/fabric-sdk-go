package dconfig

import (
	"encoding/json"
	"io/ioutil"

	"../derrors"
	"../tools"
)

func init() {
	initValues()
}

func SetFileNameKey(flag, key string, defaultValue string) (err error) {
	globalFileKey = key
	return Register(flag, key, defaultValue, "Config file name")
}

func Register(flag, key string, defaultValue interface{}, usage string) (err error) {
	key = tools.STDString(key)
	if len(flag) > 0 {
		one, ok := globalFlags[key]
		if ok {
			if one.Name == key && one.Short == flag && one.Value == defaultValue && one.Usage == usage {
				return nil
			}
			return derrors.ErrorSameKeyExistf(key)
		}
		globalFlags[key] = flagsPair{
			Name:  key,
			Short: flag,
			Value: defaultValue,
			Usage: usage,
		}
	}

	one, _ := globalContainer[key]
	one.Default = defaultValue
	globalContainer[key] = one

	return nil
}

func Get(key string) (value interface{}, ok bool) {
	key = tools.STDString(key)
	value, ok = globalValues[key]
	return
}

func GetStringByKey(key string) string {
	v, ok := Get(key)
	if !ok {
		return ""
	}
	return GetString(v, "")
}

func GetBoolByKey(key string) bool {
	v, ok := Get(key)
	if !ok {
		return false
	}
	return GetBool(v)
}

func GetIntByKey(key string) int {
	v, ok := Get(key)
	if !ok {
		return 0
	}
	return GetInt(v)
}

func GetFloatByKey(key string) float64 {
	v, ok := Get(key)
	if !ok {
		return 0
	}
	return GetFloat(v)
}

func Set(key string, value interface{}) (err error) {
	key = tools.STDString(key)
	globalValues[key] = value
	return nil
}

func Flush(name string) (err error) {
	name = tools.STDPath(name)
	buff, err := json.Marshal(&globalValues)
	if err != nil {
		return
	}

	// Todo crypto

	err = ioutil.WriteFile(name, buff, 0644)
	return
}
