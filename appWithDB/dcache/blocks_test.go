package dcache

import (
	"bytes"
	"testing"

	"../fabclient"
)

func Test_Blocks(t *testing.T) {
	input := []*fabclient.MiddleCommonBlock{
		{"", 123, []byte("PreviousHash123"), []byte("DataHash123"), nil},
		{"", 456, []byte("PreviousHash456"), []byte("DataHash456"), nil},
		{"", 789, []byte("PreviousHash789"), []byte("DataHash789"), nil},
		{"", 234, []byte("PreviousHash234"), []byte("DataHash234"), nil},
	}
	last := len(input) - 1

	err := PushBlock(input[0])
	if err != nil {
		t.Fatal(err)
	}

	one, err := FetchBlock(input[0].Number)
	if err != nil {
		t.Fatal(err)
	}

	if one == nil || 0 != bytes.Compare(one.DataHash, input[0].DataHash) {
		t.Fatal("fetch wrong")
	}

	err = PushBlockMore(input)
	if err != nil {
		t.Fatal(err)
	}

	one, err = FetchBlockHash(input[last].DataHash)
	if err != nil {
		t.Fatal(err)
	}

	if one == nil || 0 != bytes.Compare(one.PreviousHash, input[last].PreviousHash) {
		t.Fatal("fetch wrong")
	}
}
