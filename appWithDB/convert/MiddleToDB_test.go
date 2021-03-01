package convert

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_one(t *testing.T) {

	buff := []byte{0, 0x31, 0x32, 0x33, 0, 0x41, 0x42, 0x43, 0}

	one, err := json.Marshal(string(buff))

	fmt.Println(string(one), err)
}
