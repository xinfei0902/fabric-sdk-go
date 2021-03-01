package chaincode

import (
	"fmt"
	"net/http"

	"crypto/sha256"
	"static/asymmetric"
	gm "static/gm"
	sm2 "static/gm/sm2"
	"static/symmetric"

	"../../dcache"
	"../../derrors"
	"../../djson"
	"../../fabclient"
	"../../tools"
	"../../web"
	"github.com/pkg/errors"
)

// CCByAddressMake for normal handle makes
type CCByAddressMake struct {
	path      string
	invoke    bool
	chaincode string
	function  string
	document  string
}

//SignatureInfo 用户信息
type SignatureInfo struct {
	Address    string `json:"address"`
	RandomCode []byte `json:"randomCode"`
	Signature  []byte `json:"signature"`
}

//To out address
type To struct {
	ToAddress string  `json:"address"`
	Symbol    string  `json:"symbol"`
	Amount    float64 `json:"amount"`
}

//Transaction 交易信息
type Transaction struct {
	FromAddress string `json:"address"`
	UserPwd     string `json:"userPwd"`
	Symbol      string `json:"symbol"`
	Desc        string `json:"desc"`
	To          []To   `json:"to"`
}

// NewCCByAddressMake create one make
func NewCCByAddressMake(path string, invoke bool, chaincode string, function string, document string) *CCByAddressMake {
	return &CCByAddressMake{path, invoke, chaincode, function, document}
}

// Path for web.ServiceHandle
func (make *CCByAddressMake) Path() string {
	return make.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (make *CCByAddressMake) MakeChainCodeHandle() web.ChainCodeHandle {
	return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
		if len(body) == 0 {
			err = derrors.ErrorEmptyValue
			return
		}
		//Unmarshal request body
		var signatureInfo SignatureInfo
		err = djson.Unmarshal(body, &signatureInfo)
		if err != nil {
			err = errors.WithMessage(err, "Parse input value")
			return
		}
		privBase64 := dcache.GetPrivWithAddress(signatureInfo.Address)
		if privBase64 == "" {
			err = derrors.ErrorPrivKey
			return
		}
		privPEM, err := tools.Base64Decode([]byte(privBase64))
		if err != nil || len(privPEM) == 0 {
			fmt.Println(err, "Base64Decode privkey false.")
			return
		}
		privPwd := dcache.GetPrivPwdWithAddress(signatureInfo.Address)
		if privPwd == "" {
			err = derrors.ErrorPrivKey
			return
		}

		//decode to priv obj
		privKey, err := gm.PEMToPrivateKey(privPEM, []byte(privPwd))
		if err != nil {
			err = errors.WithMessage(err, "poe to priv key obj false.")
			return
		}

		//create priv with timestamp
		seed, err := asymmetric.NewSeed(string(body))
		if err != nil {
			err = errors.WithMessage(err, "new seed false.")
			return
		}
		//sign with timestamp
		//		context := []byte(body)
		opt := sha256.New()
		opt.Write(body)
		hashed := opt.Sum(nil)
		signed, err := gm.Sign(seed, privKey, hashed)
		if err != nil {
			err = errors.WithMessage(err, "sign false.")
			return
		}
		//make arg0
		signatureInfo.RandomCode = hashed
		signatureInfo.Signature = signed
		arg0, err := djson.Marshal(&signatureInfo)
		if err != nil {
			err = errors.WithMessage(err, "signatureInfo marsha1 false.")
			return
		}

		_, ok := args["async"]
		objResp, err := fabclient.CallChainCode(
			ok, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0), body})
		if err != nil {
			return "", nil, err
		}

		if objResp.Success == false {
			err = errors.New(objResp.Message)
			return
		}
		id = objResp.TxID
		switch make.function {
		case dcache.MsgForusrInfo:
			ret, err = make.UserInfo(privKey, objResp.Payload)
			if err != nil {
				err = errors.WithMessage(err, "user info query false.")
				return
			}
		case dcache.MsgForrcpInfo:
			ret, err = make.POEGetData(privKey, objResp.Payload)
			if err != nil {
				err = errors.WithMessage(err, "poe get data false.")
				return
			}
		case dcache.MsgForusrLogoff:
			err = make.RemovePrivCache(signatureInfo.Address)
			if err != nil {
				err = errors.WithMessage(err, "remove priv cache false.")
				return
			}
		default:
			fmt.Println("----------------default return value in-----------------")
			fmt.Println(objResp.Payload)
			fmt.Println(string(objResp.Payload))
			ret = objResp.Payload
			fmt.Println(ret)
			fmt.Println("----------------default return value out-----------------")
		}
		return
	}
}

// RemovePrivCache for priv remove
func (make *CCByAddressMake) RemovePrivCache(address string) (err error) {
	if address == "" {
		err = errors.WithMessage(err, "address is empty.")
		return
	}
	dcache.RemoveCacheData(address)
	return nil
}

// UserInfo for chaincode userquery
func (make *CCByAddressMake) UserInfo(priv *sm2.PrivateKey, payload []byte) (ret []byte, err error) {
	if len(payload) == 0 {
		err = errors.WithMessage(err, "payload is empty.")
		return
	}
	//Unmarshal response body
	var identity Identity
	err = djson.Unmarshal(payload, &identity)
	if err != nil {
		err = errors.WithMessage(err, "Unmarshal payload false")
		return
	}
	name, err := tools.Base64Decode([]byte(identity.Name))
	if err != nil {
		err = errors.WithMessage(err, "Base64Decode name false")
		return
	}
	uniqueID, err := tools.Base64Decode([]byte(identity.UniqueID))
	if err != nil {
		err = errors.WithMessage(err, "Base64Decode id false")
		return
	}
	userName, err := tools.Base64Decode([]byte(identity.UserName))
	if err != nil {
		err = errors.WithMessage(err, "Base64Decode id false")
		return
	}

	//crypt user priv info use sm4 with user public key
	pubByte, err := gm.PublicKeyToPEM(&priv.PublicKey)
	if err != nil || len(pubByte) == 0 {
		err = errors.WithMessage(err, "PublicKeyToPEM false.")
		return
	}
	pubmd5, err := tools.SumHashMD5(string(pubByte))
	if err != nil || len(pubmd5) != 16 {
		err = errors.WithMessage(err, "hash md5 false ,size error.")
		return
	}

	block, err := gm.NewSM4Cipher(pubmd5)
	if err != nil {
		err = errors.WithMessage(err, "NewSM4Cipher false.")
		return
	}
	name, err = symmetric.DecodeByBlock(block, nil, name)
	if err != nil {
		err = errors.WithMessage(err, "encode IDInfo false.")
		return
	}
	uniqueID, err = symmetric.DecodeByBlock(block, nil, uniqueID)
	if err != nil {
		err = errors.WithMessage(err, "encode IDInfo false.")
		return
	}
	userName, err = symmetric.DecodeByBlock(block, nil, userName)
	if err != nil {
		err = errors.WithMessage(err, "encode userName false.")
		return
	}
	identity.Name = string(name)
	identity.UniqueID = string(uniqueID)
	identity.UserName = string(userName)
	ret, err = djson.Marshal(identity)
	if err != nil {
		err = errors.WithMessage(err, "marshal user info false.")
		return
	}
	return ret, err
}

// POEGetData for chaincode userquery
func (make *CCByAddressMake) POEGetData(priv *sm2.PrivateKey, payload []byte) (ret []byte, err error) {
	if len(payload) == 0 {
		err = errors.WithMessage(err, "payload is empty.")
		return
	}
	//Unmarshal response body
	var ledger Ledger
	err = djson.Unmarshal(payload, &ledger)
	if err != nil {
		err = errors.WithMessage(err, "Unmarshal payload false")
		return
	}
	if ledger.IsDecode == false {
		return payload, nil
	}
	data, err := tools.Base64Decode([]byte(ledger.Data))
	if err != nil {
		err = errors.WithMessage(err, "Base64Decode data false")
		return nil, err
	}
	pubmd5, err := tools.SumHashMD5(dcache.POEDataEncodePwd)
	if err != nil || len(pubmd5) != 16 {
		err = errors.WithMessage(err, "hash md5 false ,size error.")
		return nil, err
	}
	block, err := gm.NewSM4Cipher(pubmd5)
	if err != nil {
		err = errors.WithMessage(err, "NewSM4Cipher false.")
		return nil, err
	}
	decodedData, err := symmetric.DecodeByBlock(block, nil, data)
	if err != nil {
		err = errors.WithMessage(err, "decode data false.")
		return nil, err
	}
	ledger.Data = string(decodedData)

	ret, err = djson.Marshal(ledger)
	if err != nil {
		err = errors.WithMessage(err, "marshal user info false.")
		return nil, err
	}
	return ret, err

}

// MakeHTTPHandle for web.ServiceHandle
func (make *CCByAddressMake) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (make *CCByAddressMake) Doc() string {
	return make.document
}
