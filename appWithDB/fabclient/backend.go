package fabclient

import (
	"fmt"
	"math"

	"github.com/golang/protobuf/proto"
	cliEvent "github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	rwsetutil "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"../tools"
)

// fabric-sdk-go\internal\github.com\hyperledger\fabric\protos\gossip
// func init() {

// }

func LoopBlockEvent(cb func(*MiddleCommonBlock), quit <-chan int) (err error) {
	if cb == nil {
		panic("Shoud not release")
	}
	opt, err := mainSetup.NewEventClient(cliEvent.WithBlockEvents(),
		cliEvent.WithSeekType(seek.Newest),
		cliEvent.WithBlockNum(math.MaxUint64))
	if err != nil {
		return errors.WithMessage(err, "New event client")
	}

	blockReg, eventBlock, err := opt.RegisterBlockEvent()
	if err != nil {
		return errors.WithMessage(err, "Register block event")
	}

	defer opt.Unregister(blockReg)

	caller := func(middle *MiddleCommonBlock) {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithField("error", err).Warning("Panic occur in Block Event CallBack")
			}
		}()

		cb(middle)
	}

Loop:
	for {
		select {
		case one, ok := <-eventBlock:
			if !ok {
				break Loop
			}
			ret, err := ParserBlockEvent(one)
			if err != nil {
				continue
			}
			caller(ret)
		case <-quit:
			break Loop
		}
	}
	return nil
}

type MiddleTranNode struct {
	TxValidationString string `json:"txValidationString,omitempty"`

	ChannelID string `json:"channelID,omitempty"`
	TxID      string `json:"txid,omitempty"`
	Type      int32  `json:"type,omitempty"`
	Epoch     uint64 `json:"epoch,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Version   int32  `json:"version,omitempty"`

	Transaction []EndorserTransactionType `json:"transaction,omitempty"`

	Config []ConfigEnvelope `json:"config,omitempty"`
}

type MiddleCommonBlock struct {
	SourceURL    string `json:"sourceUrl,omitempty"`
	Number       uint64 `json:"number,omitempty"`
	PreviousHash []byte `json:"preHash,omitempty"`
	DataHash     []byte `json:"dataHash,omitempty"`
	LastConfig   uint64 `json:"lastconfig,omitempty"`

	Child []MiddleTranNode `json:"dataInfo,omitempty"`
}

type ResultReadSetPair struct {
	Key             string `json:"key,omitempty"`
	VersionTxNum    uint64 `json:"vTxNum,omitempty"`
	VersionBlockNum uint64 `json:"vBlockNum,omitempty"`
}

type ResultWriteSetPair struct {
	Key      string `json:"key,omitempty"`
	Value    []byte `json:"value,omitempty"`
	IsDelete bool   `json:"isDelete,omitempty"`
}

type ResultQuerySetPair struct {
	StartKey     string              `json:"startKey,omitempty"`
	EndKey       string              `json:"endKey,omitempty"`
	ItrExhausted bool                `json:"itrExhausted,omitempty"`
	Reads        []ResultReadSetPair `json:"reads,omitempty"`
}

type ResultPair struct {
	NameSpace string               `json:"nameSpace,omitempty"`
	Read      []ResultReadSetPair  `json:"reads,omitempty"`
	Write     []ResultWriteSetPair `json:"writes,omitempty"`
	Query     []ResultQuerySetPair `json:"querys,omitempty"`
}

type EndorserTransactionType struct {
	ChainCodeName    string `json:"ccName,omitempty"`
	ChainCodePath    string `json:"ccPatch,omitempty"`
	ChainCodeVersion string `json:"ccVersion,omitempty"`

	RespStatus  int32        `json:"respStatus,omitempty"`
	RespMsg     string       `json:"respMsg,omitempty"`
	RespPayload []byte       `json:"respPayload,omitempty"`
	Result      []ResultPair `json:"respResult,omitempty"`

	EventChainCodeID string `json:"eventccName,omitempty"`
	EventName        string `json:"eventName,omitempty"`
	EventPayload     []byte `json:"eventPayload,omitempty"`
	EventTxID        string `json:"eventTxid,omitempty"`

	ReqCCName      string            `json:"reqccName,omitempty"`
	ReqCCPath      string            `json:"reqccPath,omitempty"`
	ReqCCVersion   string            `json:"reqccVersion,omitempty"`
	ReqArgs        [][]byte          `json:"reqArgs,omitempty"`
	ReqDecorations map[string][]byte `json:"reqDecorations,omitempty"`
	ReqTimeout     int32             `json:"reqTimeout,omitempty"`
	ReqType        string            `json:"reqType,omitempty"`
}

var (
	ErrorNoData     = errors.New("no data")
	ErrorUnkownType = errors.New("unknown Type")
)

func (opt *MiddleTranNode) GetTypeName() string {
	return MiddleTranNodeTypeToString(opt.Type)
}

func MiddleTranNodeTypeToString(n int32) string {
	return cb.HeaderType_name[n]
}

func ParserEnvelopePayload(obj *cb.Envelope) (child *MiddleTranNode, err error) {

	child = &MiddleTranNode{}
	var buff []byte

	payloadMain, err := protos_utils.GetPayload(obj)
	if err != nil {
		return
	}

	header := payloadMain.GetHeader()
	if header == nil {
		err = ErrorNoData
		return
	}
	buff = header.GetChannelHeader()
	// header.GetSignatureHeader()
	// header done

	var channelheader cb.ChannelHeader
	err = proto.Unmarshal(buff, &channelheader)
	if err != nil {
		return
	}

	child.ChannelID = channelheader.GetChannelId()
	child.TxID = channelheader.GetTxId()
	child.Type = channelheader.GetType()
	child.Epoch = channelheader.GetEpoch()
	// channelheader.GetExtension()
	child.Timestamp = channelheader.GetTimestamp().GetSeconds()
	// channelheader.GetTlsCertHash()
	child.Version = channelheader.GetVersion()
	// channelheader done

	buff = payloadMain.GetData()
	// payloadMain done

	switch cb.HeaderType(child.Type) {
	case cb.HeaderType_ENDORSER_TRANSACTION:

		more, err := ParseEndorserTransaction(buff)
		if err != nil {
			return nil, err
		}
		child.Transaction = more
	case cb.HeaderType_CONFIG:
		// Todo parse config block

		one, err := ParseConfigEnvelope(buff)
		if err != nil {
			one = nil
		}

		child.Config = []ConfigEnvelope{*one}

	default:
		// fmt.Println(">> Unknown ", cb.HeaderType_ENDORSER_TRANSACTION)
		err = ErrorUnkownType
		return
	}
	return
}

func ParserBlockChildData(input []byte) (child *MiddleTranNode, err error) {
	if len(input) == 0 {
		err = ErrorNoData
		return
	}
	obj, err := protos_utils.GetEnvelopeFromBlock(input)
	if err != nil {
		return
	}

	// obj.Signature
	// obj done

	return ParserEnvelopePayload(obj)
}

func ParseEndorserTransaction(input []byte) (ret []EndorserTransactionType, err error) {
	tx, err := protos_utils.GetTransaction(input)
	if err != nil {
		return
	}

	ret = make([]EndorserTransactionType, 0, len(tx.GetActions()))

	for _, one := range tx.GetActions() {
		var hold EndorserTransactionType
		if one == nil {
			continue
		}
		// buff := one.GetHeader()
		// if len(buff) > 0 {
		// 	// signHead, err := protos_utils.GetSignatureHeader(buff)
		// 	// if err == nil && signHead != nil {
		// 	// 	fmt.Println(">>[signHead.GetNonce]", string(signHead.GetNonce()))
		// 	// 	fmt.Println(">>[signHead.GetCreator]", string(signHead.GetCreator()))
		// 	// }
		// }

		cap, err := protos_utils.GetChaincodeActionPayload(one.GetPayload())
		if err != nil {
			continue
		}

		cpp, err := protos_utils.GetChaincodeProposalPayload(cap.GetChaincodeProposalPayload())
		if err != nil {
			continue
		}
		req := pb.ChaincodeInvocationSpec{}

		err = proto.Unmarshal(cpp.GetInput(), &req)
		if err != nil {
			continue
		}
		if req.ChaincodeSpec != nil {
			hold.ReqCCName = req.ChaincodeSpec.ChaincodeId.Name
			hold.ReqCCPath = req.ChaincodeSpec.ChaincodeId.Path
			hold.ReqCCVersion = req.ChaincodeSpec.ChaincodeId.Version
			hold.ReqArgs = req.ChaincodeSpec.Input.Args
			hold.ReqDecorations = req.ChaincodeSpec.Input.Decorations
			hold.ReqTimeout = req.ChaincodeSpec.Timeout
			hold.ReqType = req.ChaincodeSpec.Type.String()
		}

		// always empty
		// for k, v := range cpp.TransientMap {
		// 	fmt.Println(k)
		// 	fmt.Println(string(v))
		// }

		// not to explore
		// if len(cap.GetAction().GetEndorsements()) > 0 {
		// 	// OrgMSP & cert PEM
		// 	fmt.Println(">>[cap.GetAction.GetEndorsements[0].GetEndorser]", string(cap.GetAction().GetEndorsements()[0].GetEndorser()))
		// 	// Signahash
		// 	fmt.Println(">>[cap.GetAction.GetEndorsements[0].Signature]", string(cap.GetAction().GetEndorsements()[0].GetSignature()))
		// }

		prp, err := protos_utils.GetProposalResponsePayload(cap.GetAction().GetProposalResponsePayload())
		// cap done

		if err != nil {
			continue
		}

		// fmt.Println(">>[prp.GetProposalHash]", string(prp.GetProposalHash()))

		chaincodeAction, err := protos_utils.GetChaincodeAction(prp.GetExtension())
		// prp done

		if err != nil {
			continue
		}

		hold.ChainCodeName = chaincodeAction.GetChaincodeId().GetName()
		hold.ChainCodePath = chaincodeAction.GetChaincodeId().GetPath()
		hold.ChainCodeVersion = chaincodeAction.GetChaincodeId().GetVersion()

		hold.RespStatus = chaincodeAction.GetResponse().GetStatus()
		hold.RespMsg = chaincodeAction.GetResponse().GetMessage()
		hold.RespPayload = chaincodeAction.GetResponse().GetPayload()
		if len(hold.RespPayload) > 0 {
			tmp, err := tools.Base64Decode(hold.RespPayload)
			if err == nil {
				hold.RespPayload = tmp
			}
		}

		txRwSet := &rwsetutil.TxRwSet{}
		err = txRwSet.FromProtoBytes(chaincodeAction.GetResults())
		if err == nil {
			hold.Result = make([]ResultPair, len(txRwSet.NsRwSets))

			for i, one := range txRwSet.NsRwSets {
				hold.Result[i].NameSpace = one.NameSpace
				if one.KvRwSet == nil {
					continue
				}
				if len(one.KvRwSet.Reads) > 0 {
					oneHolder := make([]ResultReadSetPair, 0, len(one.KvRwSet.Reads))

					for _, two := range one.KvRwSet.Reads {
						tmp := ResultReadSetPair{}
						tmp.Key = two.GetKey()
						v := two.GetVersion()
						if v != nil {
							tmp.VersionTxNum = v.GetTxNum()
							tmp.VersionBlockNum = v.GetBlockNum()
						}
						oneHolder = append(oneHolder, tmp)
					}
					if len(oneHolder) > 0 {
						hold.Result[i].Read = oneHolder
					}
				}

				if len(one.KvRwSet.Writes) > 0 {
					oneHolder := make([]ResultWriteSetPair, 0, len(one.KvRwSet.Writes))

					for _, two := range one.KvRwSet.Writes {
						tmp := ResultWriteSetPair{}
						tmp.Key = two.GetKey()
						tmp.IsDelete = two.GetIsDelete()
						tmp.Value = two.GetValue()
						oneHolder = append(oneHolder, tmp)
					}
					if len(oneHolder) > 0 {
						hold.Result[i].Write = oneHolder
					}
				}

				if len(one.KvRwSet.RangeQueriesInfo) > 0 {
					oneHolder := make([]ResultQuerySetPair, 0, len(one.KvRwSet.RangeQueriesInfo))

					for _, two := range one.KvRwSet.RangeQueriesInfo {
						tmp := ResultQuerySetPair{}
						tmp.StartKey = two.GetStartKey()
						tmp.EndKey = two.GetEndKey()
						tmp.ItrExhausted = two.GetItrExhausted()
						three := two.GetRawReads()
						if three != nil && len(three.KvReads) > 0 {
							threeHolder := make([]ResultReadSetPair, 0, len(three.KvReads))
							for _, four := range three.KvReads {
								foc := ResultReadSetPair{}
								foc.Key = four.GetKey()
								if four.GetVersion() != nil {
									foc.VersionTxNum = four.GetVersion().GetTxNum()
									foc.VersionBlockNum = four.GetVersion().GetBlockNum()
								}
								threeHolder = append(threeHolder, foc)
							}
							if len(threeHolder) > 0 {
								tmp.Reads = threeHolder
							}
						}

						oneHolder = append(oneHolder, tmp)
					}

					if len(oneHolder) > 0 {
						hold.Result[i].Query = oneHolder
					}
				}

			}

		}

		// chaincodeAction done
		ccEvent, err := protos_utils.GetChaincodeEvents(chaincodeAction.GetEvents())
		if err == nil && ccEvent != nil {
			hold.EventChainCodeID = ccEvent.GetChaincodeId()
			hold.EventName = ccEvent.GetEventName()
			hold.EventPayload = ccEvent.GetPayload()
			hold.EventTxID = ccEvent.GetTxId()
		}

		ret = append(ret, hold)
	}

	return
}

func ParserBlockEvent(input *fab.BlockEvent) (ret *MiddleCommonBlock, err error) {
	ret, err = ParserBlockData(input.Block)
	if err != nil {
		return
	}
	ret.SourceURL = input.SourceURL
	return
}

func ParserBlockData(input *cb.Block) (ret *MiddleCommonBlock, err error) {
	ret = &MiddleCommonBlock{}

	mainHeader := input.GetHeader()
	if mainHeader != nil {
		ret.Number = mainHeader.GetNumber()
		ret.PreviousHash = mainHeader.GetPreviousHash()
		ret.DataHash = mainHeader.GetDataHash()
	}

	dataOpt := input.GetData()
	if dataOpt == nil || len(dataOpt.GetData()) == 0 {
		err = ErrorNoData
		return
	}

	ret.Child = make([]MiddleTranNode, 0, len(dataOpt.GetData()))
	for i, one := range dataOpt.GetData() {
		child, err := ParserBlockChildData(one)
		if err != nil {
			if err == ErrorNoData {
				continue
			}
			err = errors.WithMessage(err, fmt.Sprintf("Parse child No.%d", i))
			return nil, err
		}

		ret.Child = append(ret.Child, *child)
	}

	metaOpt := input.GetMetadata()
	if metaOpt == nil {
		return ret, nil
	}
	txValidationFlags := metaOpt.Metadata[cb.BlockMetadataIndex_TRANSACTIONS_FILTER]
	if len(txValidationFlags) != len(ret.Child) {
		return ret, nil
	}

	for i := range ret.Child {
		name, ok := pb.TxValidationCode_name[int32(txValidationFlags[i])]
		if !ok {
			continue
		}
		ret.Child[i].TxValidationString = name
	}

	one, err := GetLastConfigIndexFromBlock(input)
	if err == nil {
		ret.LastConfig = one
	}

	return ret, nil
}

// GetLastConfigIndexFromBlock retrieves the index of the last config block as encoded in the block metadata
func GetLastConfigIndexFromBlock(block *cb.Block) (uint64, error) {
	md, err := GetMetadataFromBlock(block, cb.BlockMetadataIndex_LAST_CONFIG)
	if err != nil {
		return 0, err
	}
	lc := &cb.LastConfig{}
	err = proto.Unmarshal(md.Value, lc)
	if err != nil {
		return 0, err
	}
	return lc.Index, nil
}

// GetMetadataFromBlock retrieves metadata at the specified index.
func GetMetadataFromBlock(block *cb.Block, index cb.BlockMetadataIndex) (*cb.Metadata, error) {
	md := &cb.Metadata{}
	err := proto.Unmarshal(block.Metadata.Metadata[index], md)
	if err != nil {
		return nil, err
	}
	return md, nil
}
