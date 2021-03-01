package dcache

import (
	"testing"

	"../fabclient"
)

func Test_peers(t *testing.T) {
	input := []fabclient.SystemInformationStruct{
		{123, "peer0.org0.com:7051", 123, []byte("PreviousBlockHash"), []byte("CurrentBlockHash"), ""},
		{123, "peer0.org1.com:7051", 456, []byte("PreviousBlockHash"), []byte("CurrentBlockHash"), ""},
		{123, "peer0.org2.com:7051", 789, []byte("PreviousBlockHash"), []byte("CurrentBlockHash"), ""},
	}
	err := SetPeers(input)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := GetPeers([]string{"peer0.org0.com:7051"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ret) != 1 {
		t.Fatal("count")
	}

	ret, err = GetPeers([]string{"peer0.org0.com:7051", "peer0.org2.com:7051"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ret) != 2 {
		t.Fatal("count")
	}

	ret, err = GetPeers([]string{"peer0.org0.com:7051", "peer0.org1.com:7051", "peer0.org2.com:7051"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ret) != 3 {
		t.Fatal("count")
	}
}
