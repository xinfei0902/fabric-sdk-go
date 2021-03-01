package fabclient

import (
	"encoding/hex"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"

	"../tools"
)

type SystemInformationStruct struct {
	Status            int32  `json:"-"`
	PeerAddress       string `json:"peerAddress"`
	Height            uint64 `json:"height"`
	PreviousBlockHash []byte `json:"previousBlockHash,omitempty"`
	CurrentBlockHash  []byte `json:"currentBlockHash,omitempty"`
	Message           string `json:"msg,omitempty"`
}

func QueryPeersTargets(orgs []string) (targets []string) {
	if len(orgs) > 0 {
		targets, _ = getTargetsByOrg(mainSetup.AllTargets, orgs)
		return
	}
	all := getAllPeersFromConfig()
	targets = make([]string, 0, len(all))
	for _, l := range all {
		if len(l) == 0 {
			continue
		}
		targets = append(targets, l...)
	}
	return
}

func QuerySystemInfomationFromPeer(targets []string) (ret []SystemInformationStruct, err error) {
	if len(targets) == 0 {
		err = ErrorNoTargetPeers
		return
	}

	ledgerClient, err := mainSetup.NewWholeLedgerClient()
	if err != nil {
		return
	}

	ret = make([]SystemInformationStruct, 0, len(targets))
	for _, one := range targets {
		var resp SystemInformationStruct

		bciBeforeTx, err := ledgerClient.QueryInfo(ledger.WithTargetEndpoints(one))
		if err != nil {
			resp.PeerAddress = one
			resp.Message = err.Error()
		} else {
			resp.Status = bciBeforeTx.Status
			resp.PeerAddress = bciBeforeTx.Endorser
			if bciBeforeTx.BCI != nil {
				resp.Height = bciBeforeTx.BCI.Height
				resp.PreviousBlockHash = bciBeforeTx.BCI.PreviousBlockHash
				resp.CurrentBlockHash = bciBeforeTx.BCI.CurrentBlockHash
			}
		}

		ret = append(ret, resp)
	}
	return
}

func QuerySystemInfomationEx(OrgsName []string) (ret []SystemInformationStruct, err error) {
	var targets []string
	if len(OrgsName) > 0 {
		targets, err = getTargetsByOrg(mainSetup.AllTargets, OrgsName)
		if err != nil {
			targets = nil
		}
	}

	if len(targets) == 0 {
		targets = getPeersFromConfig()
	}

	return QuerySystemInfomationFromPeer(targets)
}

func QueryBlockByHeight(Height uint64) (middle *MiddleCommonBlock, err error) {
	ledgerClient, err := mainSetup.NewSingleLedgerClient()
	if err != nil {
		return
	}
	block, err := ledgerClient.QueryBlock(Height)
	if err != nil {
		return
	}
	middle, err = ParserBlockData(block)
	if err != nil {
		return
	}
	return
}

func QueryBlockByKV(key, value string) (middle *MiddleCommonBlock, err error) {
	key = stdstring(key)

	ledgerClient, err := mainSetup.NewSingleLedgerClient()
	if err != nil {
		return
	}

	var block *common.Block

	switch key {
	case "txid":
		block, err = ledgerClient.QueryBlockByTxID(fab.TransactionID(value))
	case "number":
		var i uint64
		i, err = tools.StringToUInt64(value)
		if err != nil {
			err = errors.WithMessage(err, "unknown value")
			return
		}

		block, err = ledgerClient.QueryBlock(i)
	case "hash":
		var hash []byte
		hash, err = hex.DecodeString(value)
		if err != nil {
			err = errors.WithMessage(err, "unknown value")
			return
		}
		block, err = ledgerClient.QueryBlockByHash(hash)
	default:
		err = ErrorUnknownKey
	}

	if err != nil {
		return
	}

	middle, err = ParserBlockData(block)
	return
}

func QueryTransaction(txid string) (validation int32, ret *MiddleTranNode, err error) {
	ledgerClient, err := mainSetup.NewSingleLedgerClient()
	if err != nil {
		return
	}
	proc, err := ledgerClient.QueryTransaction(fab.TransactionID(txid))
	if err != nil {
		return
	}
	validation = proc.GetValidationCode()

	t := proc.GetTransactionEnvelope()
	if t != nil {
		// not explore
		// ret.Signature = t.GetSignature()

		ret, err = ParserEnvelopePayload(t)
	}

	return
}

func QueryChannel() (ret map[string][]string, err error) {
	reqCtx, cancel, err := mainSetup.NewReqContext()
	if err != nil {
		return
	}
	defer cancel()

	peers := getPeersFromConfig()

	target, err := mainSetup.GetProposalProcessors(peers)
	if err != nil {
		return
	}
	if len(target) == 0 {
		err = ErrorSystemResourcePeer
		return
	}

	ret = make(map[string][]string, len(target))
	for i, one := range target {

		channelQueryResponse, err := resource.QueryChannels(reqCtx, one, resource.WithRetry(retry.DefaultResMgmtOpts))
		if err != nil {
			continue
		}

		a := channelQueryResponse.GetChannels()

		list := make([]string, len(a))
		for j, two := range a {
			list[j] = two.GetChannelId()
		}

		ret[peers[i]] = list
	}
	return
}

type ChainCodeListStruct struct {
	ID      string `json:"ID"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Version string `json:"version"`
	Escc    string `json:"-"`
	Vscc    string `json:"-"`
	Input   string `json:"-"`
}

func QueryChainCodeList() (ret map[string][]ChainCodeListStruct, err error) {

	reqCtx, cancel, err := mainSetup.NewReqContext()
	if err != nil {
		return
	}
	defer cancel()

	peers := getPeersFromConfig()

	target, err := mainSetup.GetProposalProcessors(peers)
	if err != nil {
		return
	}
	if len(target) == 0 {
		err = ErrorSystemResourcePeer
		return
	}

	ret = make(map[string][]ChainCodeListStruct, len(target))
	for i, one := range target {

		chaincodeQueryResponse, err := resource.QueryInstalledChaincodes(reqCtx, one, resource.WithRetry(retry.DefaultResMgmtOpts))
		if err != nil {
			continue
		}

		a := chaincodeQueryResponse.Chaincodes

		list := make([]ChainCodeListStruct, len(a))
		for i, chaincode := range chaincodeQueryResponse.Chaincodes {
			list[i].ID = hex.EncodeToString(chaincode.GetId())
			list[i].Name = chaincode.GetName()
			list[i].Path = chaincode.GetPath()
			list[i].Version = chaincode.GetVersion()
			list[i].Escc = chaincode.GetEscc()
			list[i].Vscc = chaincode.GetVscc()
			list[i].Input = chaincode.GetInput()
		}

		ret[peers[i]] = list
	}

	return
}
