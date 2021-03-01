package chaincode

import (
	"restapi/object"
	"restapi/protocal"

	"github.com/pkg/errors"

	"encoding/json"
	"net/http"

	"../../derrors"
	"../../fabclient"
	"../../web"

	cross "restapi/protocal"
)

// Output

type FabricAccountPair struct {
	TxID    string           `json:"txid"`
	Account *FabricAssetPair `json:"account,omitempty"`
}

type BitcoinTranscations struct {
	TxID string `json:"txid,omitempty"`

	Balances []object.Asset `json:"balances,omitempty"`
}

type CrossChainV1Output struct {
	Bitcoin *BitcoinTranscations `json:"bitcoin"`
	Fabric  *FabricAccountPair   `json:"fabric"`
}

// Input

type FabricAssetPair struct {
	Address string  `json:"address"`
	Count   float64 `json:"count"`
}

type BitcoinAssetPair struct {
	TxID string `json:"txid"`

	Channel string  `json:"channel"`
	Address string  `json:"address"`
	Asset   string  `json:"asset"`
	Count   float64 `json:"qty"`
}

// CrossChainV1Input request object
// fabric account
// bitcoin account
// trade 1: f -> b; -1: b -> f; 0: no change data
type CrossChainV1Input struct {
	FabricAddress *FabricAssetPair  `json:"fabric,omitempty"`
	BitAddress    *BitcoinAssetPair `json:"bitcoin,omitempty"`
	Trade         int               `json:"trade,omitempty"`
}

// CrossChainV1 for normal handle makes
type CrossChainV1 struct {
	path     string
	node     string
	document string
}

// NewCrossChainV1 create one make
func NewCrossChainV1(path string, node, document string) *CrossChainV1 {
	return &CrossChainV1{path, node, document}
}

// Path for web.ServiceHandle
func (opt *CrossChainV1) Path() string {
	return opt.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (opt *CrossChainV1) MakeChainCodeHandle() web.ChainCodeHandle {
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

		if params.FabricAddress != nil && params.BitAddress != nil &&
			params.FabricAddress.Count > 0 && params.BitAddress.Count > 0 &&
			len(params.FabricAddress.Address) > 0 && len(params.BitAddress.Address) > 0 {
			if params.Trade > 0 {
				ret, err = opt.moveFabricToBitcoinV1(params.FabricAddress, params.BitAddress)
				return
			}

			if params.Trade < 0 {
				ret, err = opt.moveBitcoinToFabricV1(params.BitAddress, params.FabricAddress)
				return
			}
		}

		one := &CrossChainV1Output{}
		if params.FabricAddress != nil && len(params.FabricAddress.Address) > 0 {
			one.Fabric, err = fetchFabric(params.FabricAddress, opt.node)
			if err != nil {
				err = errors.WithMessage(err, "Get fabric account error")
				return
			}
		}
		if params.BitAddress != nil && len(params.BitAddress.Address) > 0 {
			one.Bitcoin, err = fetchBitcoin(params.BitAddress)
			if err != nil {
				err = errors.WithMessage(err, "Get fabric account error")
				return
			}
		}

		ret = one

		return
	}
}

// MakeHTTPHandle for web.ServiceHandle
func (opt *CrossChainV1) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (opt *CrossChainV1) Doc() string {
	return opt.document
}

func (opt *CrossChainV1) moveFabricToBitcoinV1(from *FabricAssetPair, to *BitcoinAssetPair) (ret *CrossChainV1Output, err error) {
	if from == nil || to == nil {
		err = derrors.ErrorEmptyValue
		return
	}
	var buff []byte

	if from.Count > 0 {
		buff, err = json.Marshal(from)
		if err != nil {
			err = errors.WithMessage(err, "Middle json tranlation")
			return
		}
	}

	ret = &CrossChainV1Output{}

	// reduce fabric
	if len(buff) > 0 {
		objResp, err := fabclient.CallChainCode(false, true, opt.node, "reduce", [][]byte{buff})
		if err != nil {
			err = errors.WithMessage(err, "Reduce fabric account error")
			return nil, err
		}

		if objResp.Success == false {
			err = errors.New(objResp.Message)
			return nil, err
		}

		ret.Fabric = outputFabric(objResp)

	}

	// add bitcoin

	resp, err := cross.SendAsset(to.Channel, to.Address, to.Asset, to.Count, 0, "", "")
	if err != nil {
		if len(buff) > 0 {
			objResp, err1 := fabclient.CallChainCode(true, true, opt.node, "increase", [][]byte{buff})
			if err1 != nil {
				err = errors.WithMessage(err1, "Recover fabric account error after send asset failed: "+err.Error())
				return nil, err
			}

			if objResp.Success == false {
				err = errors.New(objResp.Message)
				return nil, err
			}
		}
		err = errors.WithMessage(err, "send asset failed")
		return nil, err
	}
	ret.Bitcoin = &BitcoinTranscations{
		TxID: resp,
	}

	return
}

func (opt *CrossChainV1) moveBitcoinToFabricV1(from *BitcoinAssetPair, to *FabricAssetPair) (ret *CrossChainV1Output, err error) {
	// check bit txid
	addresses := make([]string, 0, 16)
	if len(from.TxID) > 0 {
		resp, err := protocal.GetAddresses(from.Channel, true)
		if err != nil {
			err = errors.WithMessage(err, "get address failed")
			return nil, err
		}

		list := make([]object.AddressWhole, 0, 16)
		err = json.Unmarshal([]byte(resp), &list)
		if err != nil {
			err = errors.WithMessage(err, "get address failed")
			return nil, err
		}
		for _, one := range list {
			if one.Ismine == false || one.Iswatchonly == true || len(one.Address) == 0 {
				continue
			}
			addresses = append(addresses, one.Address)
		}

		if len(addresses) == 0 {
			return nil, errors.New("get address emtpy")
		}
		// increase fabric
		resp, err = protocal.GetWalletTransaction(from.Channel, from.TxID, false, false)
		if err != nil {
			err = errors.WithMessage(err, "check txid")
			return nil, err
		}

		wt := &object.WalletTranscation{}

		err = json.Unmarshal([]byte(resp), wt)
		if err != nil {
			err = errors.WithMessage(err, "check txid")
			return nil, err
		}

		if len(wt.Myaddresses) == 0 {
			return nil, errors.New("txid not mine")
		}
		for _, one := range wt.Myaddresses {
			b := false
			for _, two := range addresses {
				if two == one {
					b = true
					break
				}
			}
			if b == false {
				return nil, errors.New("txid not mine")
			}
		}

		b := false
		for _, one := range wt.Balance.Assets {
			if one.Name == from.Asset && one.QTY == from.Count {
				b = true
				break
			}
		}
		if b == false {
			return nil, errors.New("txid asset not match")
		}
	}

	ret.Fabric, err = increasefabric(to, opt.node)
	if err != nil {
		ret.Fabric = nil
		err = nil
	}

	ret.Bitcoin, err = fetchBitcoin(from)
	if err != nil {
		ret.Bitcoin = nil
		err = nil
	}

	return
}
