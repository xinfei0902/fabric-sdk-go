package fabclient

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

var globalTransientData = []byte("Transient data in move funds...")

var globalTransientDataMap = map[string][]byte{
	"result": globalTransientData,
}

type ResponseInvokeStruct struct {
	TxID    string
	Success bool
	Message string
	Payload []byte
}

func stdResponse(response channel.Response) (resp ResponseInvokeStruct) {
	if response.ChaincodeStatus != 200 {
		resp.Success = false
		resp.Message = fmt.Sprintf("Status: %d; Error: %s", response.ChaincodeStatus, string(response.Payload))
		resp.TxID = string(response.TransactionID)
		return
	}

	resp.Success = true
	resp.TxID = string(response.TransactionID)
	resp.Message = ""
	resp.Payload = response.Payload

	if len(response.Payload) > 0 {
		if response.Payload[0] != '{' {
			tmp, err := base64.StdEncoding.DecodeString(string(response.Payload))
			if err == nil {
				resp.Payload = tmp
			}
		}
	}

	return
}

func CallChainCode(async bool, invoke bool, chainCodeID string, function string, args [][]byte, options ...channel.RequestOption) (resp ResponseInvokeStruct, err error) {
	chClient, err := mainSetup.NewChannelClient()
	if err != nil {
		return
	}

	request := channel.Request{
		ChaincodeID:  chainCodeID,
		Fcn:          function,
		Args:         args,
		TransientMap: globalTransientDataMap,
	}

	var response channel.Response

	switch {
	case false == invoke:
		// response, err = chClient.Query(request, options...)
		response, err = MyQuery(chClient, request, options...)
	case invoke:
		// response, err = chClient.Execute(request, options...)
		response, err = MyExecute(chClient, async, request, options...)
	}

	if err != nil {
		switch err.(type) {
		case multi.Errors:
			realErr, _ := err.(multi.Errors)
			for _, one := range realErr {
				if one == nil {
					continue
				}
				err = one
				return
			}
		case *status.Status:
			// realErr, _ := err.(*status.Status)
			// fmt.Println(realErr.Code)
			// fmt.Println(realErr.Message)
		default:
		}
		return
	}

	return stdResponse(response), nil
}

func InvokeChainCode(chainCodeID string, args [][]byte) (resp ResponseInvokeStruct, err error) {
	return CallChainCode(false, true, chainCodeID, "invoke", args)
}

func QueryChainCode(chainCodeID string, args [][]byte) (resp ResponseInvokeStruct, err error) {
	return CallChainCode(false, false, chainCodeID, "query", args)
}

type staticOnlyMSP struct {
}

func (*staticOnlyMSP) Accept(peer fab.Peer) bool {
	return peer.MSPID() == mainSetup.PeerOrgID || peer.MSPID() == mainSetup.PeerOrgID+"MSP"
}

var pointFilter = &staticOnlyMSP{}

func QueryChainCodeSingleOrg(chainCodeID string, args [][]byte) (resp ResponseInvokeStruct, err error) {
	return CallChainCode(false, false, chainCodeID, "query", args, channel.WithTargetFilter(pointFilter))
}
