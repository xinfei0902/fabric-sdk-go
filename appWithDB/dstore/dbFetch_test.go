package dstore

import (
	"fmt"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func Test_FetchBlockTxID(t *testing.T) {
	err := newDBClientPool("postgres", "host=192.168.0.150 port=5432 user=sinochain dbname=sinochain sslmode=disable password=sinochain ")
	if err != nil {
		t.Fatal(err)
	}
	globalDBOpt = globalDBOpt.Debug()

	ret, err := FetchBlockTxID("8c40ac6d811a17dfc2433362df44663bf0d6479d69a1f140fabe5dd07c5ccd70")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ret.DataHash)
	fmt.Println(len(ret.Transactions))
	fmt.Println(len(ret.Transactions[0].Transaction))
	fmt.Println(len(ret.Transactions[0].Events))
	fmt.Println(len(ret.Transactions[0].Actions))
}

func Test_FetchBlockRange(t *testing.T) {
	err := newDBClientPool("postgres", "host=192.168.0.150 port=5432 user=sinochain dbname=sinochain sslmode=disable password=sinochain ")
	if err != nil {
		t.Fatal(err)
	}
	globalDBOpt = globalDBOpt.Debug()

	ret, err := FetchBlockRange(2, 5)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(len(ret))
	fmt.Println(len(ret[0].Transactions))
	fmt.Println(len(ret[0].Transactions[0].Transaction))
	fmt.Println(len(ret[0].Transactions[0].Events))
	fmt.Println(len(ret[0].Transactions[0].Actions))
}

func Test_FetchTranscations(t *testing.T) {
	err := newDBClientPool("postgres", "host=192.168.0.150 port=5432 user=sinochain dbname=sinochain sslmode=disable password=sinochain ")
	if err != nil {
		t.Fatal(err)
	}
	globalDBOpt = globalDBOpt.Debug()

	ret, err := FetchTranscations("8c40ac6d811a17dfc2433362df44663bf0d6479d69a1f140fabe5dd07c5ccd70")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ret.Timestamp)
	fmt.Println(len(ret.Transaction))
	fmt.Println(len(ret.Events))
	fmt.Println(len(ret.Actions))
}
