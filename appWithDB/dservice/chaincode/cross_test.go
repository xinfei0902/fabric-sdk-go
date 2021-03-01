package chaincode

import (
	"encoding/json"
	"fmt"
	"restapi/object"
	"testing"
)

func Test_aa(t *testing.T) {
	a := 1 + 1
	fmt.Println("here", a)

	var aaa []string
	aaa = nil

	b := make([]string, 0, 5)
	b = append(b, aaa...)
	fmt.Println("here", b)
}

func Test_json(t *testing.T) {
	type One struct {
		A string
		B string
	}

	buff := []byte(`{"A": "aa", "B": "bb", "C": "cc"}`)
	one := &One{}
	err := json.Unmarshal(buff, one)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(one)

	buff = []byte(`{"A": "aa"}`)
	one = &One{}
	err = json.Unmarshal(buff, one)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(one)

	buff = []byte(`{"D": "dd"}`)
	one = &One{}
	err = json.Unmarshal(buff, one)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(one)

	buff = []byte(`{"D": "dd", "E": "ee"}`)
	one = &One{}
	err = json.Unmarshal(buff, one)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(one)
}

func Test_object_parse(t *testing.T) {
	buff := `{"result":null,"error":{"code":-705,"message":"Asset or stream with this name already exists"},"id":1}`

	// var txid string
	ret, err := object.ParseResponse(buff, nil)
	fmt.Println(err)
	fmt.Println(ret)
}
