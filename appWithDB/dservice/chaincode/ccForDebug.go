package chaincode

import (
	"fmt"
	"net/http"

	"crypto/sha256"
	"static/asymmetric"
	gm "static/gm"

	"../../dcache"
	"../../derrors"
	"../../djson"
	"../../fabclient"
	"../../tools"
	"../../web"
	"github.com/pkg/errors"
)

// CCForDebugMake for normal handle makes
type CCForDebugMake struct {
	path      string
	invoke    bool
	chaincode string
	function  string
	document  string
}

// NewCCForDebugMake create one make
func NewCCForDebugMake(path string, invoke bool, chaincode string, function string, document string) *CCForDebugMake {
	return &CCForDebugMake{path, invoke, chaincode, function, document}
}

// Path for web.ServiceHandle
func (make *CCForDebugMake) Path() string {
	return make.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (make *CCForDebugMake) MakeChainCodeHandle() web.ChainCodeHandle {
	return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
		if len(body) == 0 {
			err = derrors.ErrorEmptyValue
			return
		}

		switch make.function {
		//篡改交易数据，查看交易结果
		case dcache.MsgForastTestChangeValue:
			body, err = make.TestChangeValue(body)
			if err != nil {
				err = errors.WithMessage(err, "change value false.")
				return
			}
			make.function = dcache.MsgForastTrade
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

		//create priv with body
		seed, err := asymmetric.NewSeed(string(body))
		if err != nil {
			err = errors.WithMessage(err, "new seed false.")
			return
		}
		//sign with body
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
		ret = objResp.Payload
		return
	}
}

// TestChangeValue for web.ServiceHandle
func (make *CCForDebugMake) TestChangeValue(body []byte) (ret []byte, err error) {
	//Unmarshal request body
	var atcnt Transaction
	err = djson.Unmarshal(body, &atcnt)
	if err != nil {
		err = errors.WithMessage(err, "Parse input body false")
		return nil, err
	}
	atcnt.To[0].Amount = 1000000
	ret, err = djson.Marshal(atcnt)
	if err != nil {
		err = errors.WithMessage(err, "Marshal new value false")
		return nil, err
	}
	return ret, nil
}

// MakeHTTPHandle for web.ServiceHandle
func (make *CCForDebugMake) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (make *CCForDebugMake) Doc() string {
	return make.document
}
