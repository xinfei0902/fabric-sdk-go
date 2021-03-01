package dconfig

import (
	"encoding/json"
	"io/ioutil"
)

func ReadConfigFile(name string) (ret map[string]interface{}, err error) {
	buff, err := ioutil.ReadFile(name)
	if err != nil {
		return
	}

	// Todo crypto

	ret = make(map[string]interface{})
	err = json.Unmarshal(buff, &ret)
	return
}
