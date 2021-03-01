package chaincode

import (
	"encoding/json"
	"net/http"
	"restapi"
	"restapi/object"
	"restapi/protocal"

	"../../derrors"
	"../../fabclient"
	"../../web"
	"github.com/pkg/errors"
)

// CrossChain for normal handle makes
type CrossChain struct {
	path     string
	node     string
	kind     string
	document string
}

// NewCrossChain create one make
func NewCrossChain(path, node, kind, document string) *CrossChain {
	return &CrossChain{path, node, kind, document}
}

// Path for web.ServiceHandle
func (opt *CrossChain) Path() string {
	return opt.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (opt *CrossChain) MakeChainCodeHandle() web.ChainCodeHandle {
	switch opt.kind {
	case "issue":
		return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
			if len(body) == 0 {
				err = derrors.ErrorEmptyValue
				return
			}
			params := &CrossChainV1Input{}

			err = json.Unmarshal(body, params)
			if err != nil {
				err = errors.WithMessage(err, "Parse input value")
				return
			}

			one := &CrossChainV1Output{}

			if params.FabricAddress != nil && params.FabricAddress.Count > 0 {
				one.Fabric, err = issueFabric(params.FabricAddress, opt.node)
				if err != nil {
					err = errors.WithMessage(err, "issue fabric error")
					return
				}
			}

			if params.BitAddress != nil && params.BitAddress.Count > 0 {
				one.Bitcoin, err = issueBitCoin(params.BitAddress)
				if err != nil {
					err = errors.WithMessage(err, "issue bitcoin error")
					return
				}
			}
			ret = one
			return
		}
	case "increase":
		return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
			if len(body) == 0 {
				err = derrors.ErrorEmptyValue
				return
			}
			params := &CrossChainV1Input{}

			err = json.Unmarshal(body, params)
			if err != nil {
				err = errors.WithMessage(err, "Parse input value")
				return
			}

			one := &CrossChainV1Output{}

			if params.FabricAddress != nil {
				one.Fabric, err = increasefabric(params.FabricAddress, opt.node)
				if err != nil {
					err = errors.WithMessage(err, "issue fabric error")
					return
				}
			}

			if params.BitAddress != nil {
				one.Bitcoin, err = increaseBitCoin(params.BitAddress)
				if err != nil {
					err = errors.WithMessage(err, "issue bitcoin error")
					return
				}
			}
			ret = one
			return
		}
	case "fetch":
		return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
			if len(body) == 0 {
				err = derrors.ErrorEmptyValue
				return
			}
			params := &CrossChainV1Input{}

			err = json.Unmarshal(body, params)
			if err != nil {
				err = errors.WithMessage(err, "Parse input value")
				return
			}

			one := &CrossChainV1Output{}

			if params.FabricAddress != nil {
				one.Fabric, err = fetchFabric(params.FabricAddress, opt.node)
				if err != nil {
					err = errors.WithMessage(err, "fetch fabric error")
					return
				}
			}

			if params.BitAddress != nil {
				one.Bitcoin, err = fetchBitcoin(params.BitAddress)
				if err != nil {
					err = errors.WithMessage(err, "fetch bitcoin error")
					return
				}
			}
			ret = one
			return
		}
	}
	return nil

}

// MakeHTTPHandle for web.ServiceHandle
func (opt *CrossChain) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (opt *CrossChain) Doc() string {
	return opt.document
}

// base

func outputFabric(objResp fabclient.ResponseInvokeStruct) (ret *FabricAccountPair) {
	ret = &FabricAccountPair{
		TxID: objResp.TxID,
	}
	fap := &FabricAssetPair{}

	err := json.Unmarshal(objResp.Payload, fap)
	if err == nil {
		ret.Account = fap
	}
	return
}

func fetchFabric(input *FabricAssetPair, ccname string) (ret *FabricAccountPair, err error) {
	buff, err := json.Marshal(input)
	if err != nil {
		return
	}
	objResp, err := fabclient.CallChainCode(true, false, ccname, "fetch", [][]byte{buff})
	if err != nil {
		return
	}

	ret = outputFabric(objResp)

	return
}

func fetchBitcoin(input *BitcoinAssetPair) (ret *BitcoinTranscations, err error) {
	list, err := restapi.GetTotalBalances(input.Channel, -1, false, false)
	if err != nil {
		return
	}

	if len(input.Asset) != 0 {
		for _, one := range list {
			if input.Asset == one.Name {
				list = []object.Asset{one}
				break
			}
		}
	}

	ret = &BitcoinTranscations{
		Balances: list,
	}

	return
}

func issueFabric(input *FabricAssetPair, ccname string) (ret *FabricAccountPair, err error) {
	if len(input.Address) == 0 || input.Count <= 0 {
		err = derrors.ErrorEmptyValue
		return
	}
	buff, err := json.Marshal(input)
	if err != nil {
		return
	}
	objResp, err := fabclient.CallChainCode(false, true, ccname, "issue", [][]byte{buff})

	if err != nil {
		return
	}

	ret = outputFabric(objResp)
	return
}

func issueBitCoin(input *BitcoinAssetPair) (ret *BitcoinTranscations, err error) {
	if input.Count <= 0 && len(input.Asset) == 0 {
		return nil, derrors.ErrorEmptyValue
	}
	txid, err := restapi.Issue(input.Channel, input.Address, input.Asset, input.Count, -1, -1, "")
	if err != nil {
		return
	}

	ret = &BitcoinTranscations{
		TxID: txid,
	}
	return
}

func increasefabric(input *FabricAssetPair, ccname string) (ret *FabricAccountPair, err error) {
	buff, err := json.Marshal(input)
	if err != nil {
		return
	}
	resp, err := fabclient.CallChainCode(false, true, ccname, "increase", [][]byte{buff})
	if err != nil {
		return
	}

	ret = outputFabric(resp)
	return
}

func increaseBitCoin(input *BitcoinAssetPair) (ret *BitcoinTranscations, err error) {
	if input.Count <= 0 && len(input.Asset) == 0 {
		return nil, derrors.ErrorEmptyValue
	}
	buff, err := protocal.IssueMore(input.Channel, input.Address, input.Asset, input.Count, -1, "")
	if err != nil {
		return
	}
	ret = &BitcoinTranscations{}
	_, err = object.ParseResponse(buff, ret)
	if err != nil {
		ret = nil
	}

	return
}
