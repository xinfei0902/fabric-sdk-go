package convert

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_json(t *testing.T) {

	first := ""
	second := ""
	three := []byte{}
	one := []interface{}{
		&first,
		&UserAddrss{
			Address: &second,
		},
		&three,
	}

	buff := []byte(`["first", {"address": "bbbbbb"}]`)

	err := json.Unmarshal(buff, &one)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(first)
	fmt.Println(second)
	fmt.Println(three)

}
