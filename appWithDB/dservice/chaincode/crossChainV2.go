package chaincode

import (
	"encoding/json"
	"net/http"
	"reflect"
	"restapi"
	"restapi/object"
	"restapi/protocal"
	"time"

	"../../derrors"
	"../../fabclient"
	"../../web"
	"github.com/pkg/errors"
)

// CrossChainV2 for normal handle makes
// create stream --> stream ID, middle address
//                  | fabric - stream ID : atomic exchange hex
// A lock fabric with hash -- baaS - update stream ID status
//                                 | append atomic exchange hex with A asked
// B send exchange hex to baaS -- baaS -| 1. check hex by call complete
//                                      | 2. send hex get txid
//                                      | 3. unlock A fabric to B
//                                      | 4. update stream status
// * A offer fabric asked bitcoin
// * B offer bitcoin asked fabric
// 1. A call bitcoin create atomic exchange about asked as `A asked hex`
// 1. A call BaaS record offer and asked
//    1. BaaS call bitcoin: check `A asked hex`
//    1. BaaS call fabric: `A asked hex`, A fabric address, offer -->
//       1. chaincode: check A fabirc address, offer
//       1. chaincode: put hash(A fabric address, exchange hex) as `A asked key`: A fabric address, offer count, `A asked hex`, status as `A offer package`
//       1. chaincode: reduce A fabirc address by offer count
//       1. chaincode return `A asked key`
//    1. BaaS return `A asked key`
// 2. A cancel by `A asked key`
//    2. BaaS call fabric: `A asked key` -->
//       2. chaincode: check `A asked key` and status
//       2. chaincode: update status
//       2. chaincode: increase A fabric address by offer count
// 1. B call BaaS search by `A asked key`
//    1. BaaS call fabric: `A asked key` --> `A asked hex`, offer count
// 1. B call bitcoin complete `A asked hex`
//    1. bitcoint: `A asked hex` --> `complete hex`
// 1. B call BaaS: complete exchange, B fabric address, `A asked key`
//    1. BaaS call fabric: get `A asked key` package
//    3. check package status: complete or error or cancel --> failed
//    1. BaaS call bitcoin: decode `A asked hex` --> A asked address
//    1. BaaS call bitcoin: check `complete hex` is completed
//    2. BaaS call bitcoin: check failed
//       2. Error: Input already spent; Input: 0, txid: ..., vout: ...
//       2. check: getrawtransaction txid vout get address
//       2. if address is equal asked address try send completed
//          2. if error transaction already in block chain --> set completed mark
//          3. other failed
//       3. other failed
//    1. BaaS call fabric:
//       1. check `A asked key` status
//       1. update&add `A asked key`: status, `complete hex`, B fabric address
//    1. BaaS call bitcoin: send complete exchange --> complete txid
//    1. BaaS call fabric: `A asked key` -->
//       1. update&add status, `complete hex`, B fabric address, complete txid
//       1. increase B fabric address by offer count
//       1. return update txid
//    2. BaaS call fabric:
//       2. check `A asked key` status, `complete hex`
//       2. if same `complete hex`, B fabric address and status is running
//       2. increase B fabric address by offer count
//       2. return update txid
//    1. BaaS return update txid
type CrossChainV2 struct {
	path     string
	node     string
	kind     string
	document string
}

// NewCrossChainV2 create one make
func NewCrossChainV2(path string, node, kind, document string) *CrossChainV2 {
	return &CrossChainV2{path, node, kind, document}
}

// Path for web.ServiceHandle
func (opt *CrossChainV2) Path() string {
	return opt.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (opt *CrossChainV2) MakeChainCodeHandle() web.ChainCodeHandle {
	switch opt.kind {
	case "start":
		return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
			if len(body) == 0 {
				err = derrors.ErrorEmptyValue
				return
			}

			input := &CrossChainV2StartInput{}
			err = json.Unmarshal(body, input)
			if err != nil {
				err = errors.WithMessage(err, "Parse input value")
				return
			}

			ret, err = startRecord(input, opt.node)
			return
		}
	case "cancel":
		return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
			if len(body) == 0 {
				err = derrors.ErrorEmptyValue
				return
			}

			input := &CancelInput{}
			err = json.Unmarshal(body, input)
			if err != nil {
				err = errors.WithMessage(err, "Parse input value")
				return
			}

			hex, fid, err := fabricCancel(input.Address, input.AskedKey, opt.node)
			if err != nil {
				err = errors.WithMessage(err, "cancel fabric failed")
				return
			}

			bid, err := bitcoinDisableExchange(hex, input.BitcoinChannel)
			if err != nil {
				err = errors.WithMessage(err, "cancel bitcoin failed")
				return
			}

			one := &CrossChainV1Output{}
			one.Fabric = &FabricAccountPair{TxID: fid}
			one.Bitcoin = &BitcoinTranscations{TxID: bid}
			ret = one
			return
		}
	case "search":
		return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
			if len(body) == 0 {
				err = derrors.ErrorEmptyValue
				return
			}

			objResp, err := fabclient.CallChainCode(false, false, opt.node, "search", [][]byte{body})
			if err != nil {
				return
			}

			if objResp.Success == false {
				err = errors.New(objResp.Message)
				return
			}
			id = objResp.TxID
			ret = objResp.Payload
			return
		}
	case "complete":
		return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
			if len(body) == 0 {
				err = derrors.ErrorEmptyValue
				return
			}

			input := &CompleteInput{}
			err = json.Unmarshal(body, input)
			if err != nil {
				err = errors.WithMessage(err, "Get input failed")
				return
			}

			objResp, err := fabclient.CallChainCode(false, true, opt.node, "prepare", [][]byte{
				[]byte(input.AskedKey), []byte(input.CompleteHex), []byte(input.Address),
			})
			if err != nil {
				err = errors.WithMessage(err, "Get record failed")
				return
			}

			if objResp.Success == false {
				err = errors.New("Get record failed: " + objResp.Message)
				return
			}

			stored := &SearchOutput{}
			err = json.Unmarshal(objResp.Payload, stored)
			if err != nil {
				err = errors.WithMessage(err, "Parse record failed")
				return
			}
			switch stored.Status {
			case "running":
				ret, err = runComplete(input, stored, opt.node)
				return
			case "waiting":
				err = errors.New("prepare failed")
				return
			}
			if len(stored.Status) == 0 {
				stored.Status = "(empty)"
			}
			err = errors.New("Failed by wrong status: " + stored.Status)
			return
		}
	}
	return nil
}

// MakeHTTPHandle for web.ServiceHandle
func (opt *CrossChainV2) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (opt *CrossChainV2) Doc() string {
	return opt.document
}

//////////////////////////   start       //////////////////////////////

type CrossChainV2StartInput struct {
	AskedHex string         `json:"asked"`
	Address  string         `json:"address"`
	Offer    []object.Asset `json:"offer"`
	Asked    []object.Asset `json:"ask"`

	BitcoinChannel string `json:"channel"`

	Duration time.Duration `json:"duration"`
}

func startRecord(input *CrossChainV2StartInput, ccname string) (ret []byte, err error) {
	obj, err := restapi.DecodeRawExchange(input.BitcoinChannel, input.AskedHex, false)
	if err != nil {
		return
	}
	if obj.Complete == true || len(obj.Ask.Assets) == 0 {
		err = errors.New("exchange information wrong")
		return
	}

	// Todo here
	input.Asked = obj.Ask.Assets

	ret, err = fabricRecord(input, ccname)
	return
}

func fabricRecord(input *CrossChainV2StartInput, ccname string) (key []byte, err error) {
	buff, err := json.Marshal(input)
	if err != nil {
		return
	}
	objResp, err := fabclient.CallChainCode(false, true, ccname, "start", [][]byte{buff})
	if err != nil {
		return
	}

	if objResp.Success == false {
		err = errors.New(objResp.Message)
		return
	}

	key = objResp.Payload
	return
}

/////////////////////////   cancel     ///////////////////////////////

type CancelInput struct {
	Address  string `json:"address"`
	AskedKey string `json:"askedkey"`

	BitcoinChannel string `json:"channel"`
}

func bitcoinDisableExchange(hex string, channel string) (txid string, err error) {
	buff, err := protocal.DisableRawExchange(channel, hex)
	if err != nil {
		return
	}

	var one string

	_, err = object.ParseResponse(buff, &one)
	if err != nil {
		return
	}
	txid = one
	return
}

func fabricCancel(address, askedKey string, ccname string) (hex, txid string, err error) {
	objResp, err := fabclient.CallChainCode(false, true, ccname, "cancel", [][]byte{[]byte(address), []byte(askedKey)})
	if err != nil {
		return
	}

	if objResp.Success == false {
		err = errors.New(objResp.Message)
		return
	}
	hex = string(objResp.Payload)
	txid = objResp.TxID
	return
}

/////////////////////////// search  //////////////////////////////////////

type SearchInput struct {
	AskedKey string `json:"askedkey"`
}

type SearchOutput struct {
	Status string `json:"status"`

	// fabric
	Offer        []object.Asset `json:"offer"`
	OfferAddress string         `json:"offeraddress"`
	AskedAddress string         `json:"askedaddress"`

	// bitcoin
	BitcoinChannel string `json:"channel"`
	AskedHex       string `json:"askedhex"`
	CompleteHex    string `json:"completehex"`
	BitsendTxID    string `json:"finishedtxid"`

	Start    time.Time     `json:"start"`
	Duration time.Duration `json:"duration"`
}

/////////////////////////// complete ////////////////////////////////////

type CompleteInput struct {
	AskedKey    string `json:"askedkey"`
	CompleteHex string `json:"completehex"`
	Address     string `json:"address"`
}

type ComleteOutput struct {
	BitcoinTxID    string `json:"btid,omitempty"`
	BitcoinMessage string `json:"btmsg,omitempty"`

	FabricTxID    string `json:"fxid,omitempty"`
	FabricMessage string `json:"fmsg,omitempty"`
}

func runComplete(input *CompleteInput, stored *SearchOutput, ccname string) (ret *ComleteOutput, err error) {
	// check exchange
	completedObj, err := restapi.DecodeRawExchange(stored.BitcoinChannel, input.CompleteHex, true)
	if err != nil {
		err = errors.WithMessage(err, "Get completed exchange failed")
		return
	}

	if completedObj.Complete == false || completedObj.Cancomplete == false {
		err = errors.New("Exchange not complete")
		return
	}

	askedobj, err := restapi.DecodeRawExchange(stored.BitcoinChannel, stored.AskedHex, true)
	if err != nil {
		err = errors.WithMessage(err, "Get asked exchange failed")
		return
	}

	if len(askedobj.Exchanges) > len(completedObj.Exchanges) {
		err = errors.New("Exchange count wrong")
		return
	}

	for i, v := range askedobj.Exchanges {
		o := completedObj.Exchanges[i]
		if false == reflect.DeepEqual(&v, &o) {
			err = errors.New("Exchange not match wrong")
			return
		}
	}

	ret = &ComleteOutput{}

	// try send complete
	buff, err := protocal.SendRawTransaction(stored.BitcoinChannel, input.CompleteHex)

	var sendTxID string
	_, err = object.ParseResponse(buff, &sendTxID)
	if err != nil {
		e, ok := err.(*object.ErrorType)
		if false == ok {
			return
		}
		ret.BitcoinMessage = err.Error()
		if e.GetCode() != -27 {
			_, txid, err := fabricCancel(stored.OfferAddress, input.AskedKey, ccname)
			if err != nil {
				ret.FabricMessage = err.Error()
				return ret, nil
			}
			// cancel fabirc
			ret.FabricTxID = txid
			return ret, nil
		}
		return
	}

	objResp, err := fabclient.CallChainCode(false, true, ccname, "complete", [][]byte{
		[]byte(input.AskedKey), []byte(input.CompleteHex), []byte(input.Address), []byte(sendTxID),
	})

	if err != nil {
		ret.FabricMessage = err.Error()
		return ret, nil
	}

	ret.FabricTxID = objResp.TxID
	return ret, nil
}
