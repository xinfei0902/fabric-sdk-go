package fabclient

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	cliEvent "github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
)

func teardown() {
	CloseSetup()
}

func Test_main(t *testing.T) {
	err := InitGlobalSetup(pathvar.Subst("../../test/sinoconfig.yaml"),
		"mychannel", "Org1", "../../test/mychannel.tx", "Admin")
	if err != nil {
		t.Fatal(err)
	}

	Test_B(t)

	CloseSetup()

}

func Test_A(t *testing.T) {
	f, err := os.Create("output")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	opt, err := mainSetup.NewEventClient(cliEvent.WithBlockEvents(),
		cliEvent.WithSeekType(seek.Newest),
		cliEvent.WithBlockNum(math.MaxUint64))

	if err != nil {
		t.Fatal(err)
	}

	blockReg, eventBlock, err := opt.RegisterBlockEvent()
	if err != nil {
		t.Fatal(err)
	}

	defer opt.Unregister(blockReg)

	CCReg, eventCC, err := opt.RegisterChaincodeEvent("SIMNI", ".*")
	if err != nil {
		t.Fatal(err)
	}
	defer opt.Unregister(CCReg)

	fmt.Println("Start")
	for {
		select {
		case one, ok := <-eventBlock:
			if !ok {
				t.Fatal("closed block")
			}
			ret, err := ParserBlockData(one.Block)
			if err != nil {
				t.Fatal(err)
			}

			buff, err := json.Marshal(ret)
			if err != nil {
				t.Fatal(err)
			}
			f.WriteString(string(buff))
		case one, ok := <-eventCC:
			if !ok {
				t.Fatal("close cc")
			}
			fmt.Println("--", one.SourceURL)
			fmt.Println("--", one.TxID)
			fmt.Println("--", one.EventName)
			fmt.Println("--", one.BlockNumber)
			fmt.Println("--", one.ChaincodeID)
			fmt.Println("--", string(one.Payload))
		}
	}
}

func Test_B(t *testing.T) {
	one, err := QueryBlockByHeight(0)
	fmt.Println(err)

	buff, err := json.Marshal(one)
	fmt.Println(err)

	fmt.Println(string(buff))
}
