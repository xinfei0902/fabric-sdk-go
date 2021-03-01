package chaincode

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"crypto/sha256"
	"static/asymmetric"
	gm "static/gm"
	"static/symmetric"

	"../../dcache"
	"../../derrors"
	"../../djson"
	"../../fabclient"
	"../../tools"
	"../../web"
	"github.com/pkg/errors"
)

// CCPOEMake for normal handle makes
type CCPOEMake struct {
	path      string
	invoke    bool
	chaincode string
	function  string
	document  string
}

//Ledger 存证内容
type Ledger struct {
	IsDecode   bool    `json:"isDecode"`
	TxID       string  `json:"txid"`
	TxTimeDate string  `json:"txTimeDate"`
	Address    string  `json:"address"`
	Amount     float64 `json:"amount"`
	Hash       string  `json:"hash"`
	Data       string  `json:"data"`
	UserPwd    string  `json:"userPwd"`
	PreTxid    string  `json:"preTxid"`
}

// NewCCPOEMake create one make
func NewCCPOEMake(path string, invoke bool, chaincode string, function string, document string) *CCPOEMake {
	return &CCPOEMake{path, invoke, chaincode, function, document}
}

// Path for web.ServiceHandle
func (make *CCPOEMake) Path() string {
	return make.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (make *CCPOEMake) MakeChainCodeHandle() web.ChainCodeHandle {
	return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
		if len(body) == 0 {
			err = errors.WithMessage(err, "payload is empty.")
			return
		}
		//Unmarshal request body
		var ledger Ledger
		err = djson.Unmarshal(body, &ledger)
		if err != nil {
			err = errors.WithMessage(err, "Parse input value")
			return
		}

		var hashed []byte
		opt := sha256.New()
		if make.function == dcache.MsgForrcpBuild {
			opt.Write([]byte(ledger.Data))
			hashed = opt.Sum(nil)
			sha256 := hex.EncodeToString(hashed)
			if sha256 != ledger.Hash {
				err = errors.WithMessage(err, "data hash is not matched.")
				return
			}
		}

		pubmd5, err := tools.SumHashMD5(dcache.POEDataEncodePwd)
		if err != nil || len(pubmd5) != 16 {
			err = errors.WithMessage(err, "hash md5 false ,size error.")
			return
		}

		block, err := gm.NewSM4Cipher(pubmd5)
		if err != nil {
			err = errors.WithMessage(err, "NewSM4Cipher false.")
			return
		}
		//create priv with pwd and ID
		seed, err := asymmetric.NewSeed(ledger.Data)
		if err != nil {
			err = errors.WithMessage(err, "new seed false.")
			return
		}
		byteSeed, err := symmetric.RandomKey(block.BlockSize())
		if err != nil {
			err = errors.WithMessage(err, "read randomkey false.")
			return
		}
		_, err = io.ReadFull(seed, byteSeed)
		if err != nil {
			err = errors.WithMessage(err, "seed to byte false.")
			return
		}
		cryptedData, err := symmetric.EncodeByBlock(block, byteSeed, []byte(ledger.Data))
		if err != nil {
			err = errors.WithMessage(err, "encode IDInfo false.")
			return
		}
		opt = sha256.New()
		ledger.Data = tools.Base64Encode(cryptedData)
		opt.Write([]byte(ledger.Data))
		ledger.Hash = hex.EncodeToString(opt.Sum(nil))
		cryptBody, err := djson.Marshal(ledger)
		if err != nil {
			err = errors.WithMessage(err, "Marshal false.")
			return
		}

		//create priv with timestamp
		seed, err = asymmetric.NewSeed(string(cryptBody))
		if err != nil {
			err = errors.WithMessage(err, "new seed false.")
			return
		}

		//get the login state
		privBase64 := dcache.GetPrivWithAddress(ledger.Address)
		if privBase64 == "" {
			err = derrors.ErrorPrivKey
			return
		}

		privPEM, err := tools.Base64Decode([]byte(privBase64))
		if err != nil || len(privPEM) == 0 {
			fmt.Println(err, "Base64Decode privkey false.")
			return
		}
		privPwd := dcache.GetPrivPwdWithAddress(ledger.Address)
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

		opt = sha256.New()
		opt.Write(cryptBody)
		hashed = opt.Sum(nil)
		signed, err := gm.Sign(seed, privKey, hashed)
		if err != nil {
			err = errors.WithMessage(err, "sign false.")
			return
		}
		//make arg0
		var signatureInfo SignatureInfo
		signatureInfo.Address = ledger.Address
		signatureInfo.RandomCode = hashed
		signatureInfo.Signature = signed
		arg0, err := djson.Marshal(&signatureInfo)
		if err != nil {
			err = errors.WithMessage(err, "signatureInfo marsha1 false.")
			return
		}

		_, ok := args["async"]
		objResp, err := fabclient.CallChainCode(
			ok, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0), cryptBody})
		if err != nil {
			return "", nil, err
		}

		if objResp.Success == false {
			err = errors.New(objResp.Message)
			return
		}
		ret = objResp.Payload
		id = objResp.TxID
		return
	}
}

// MakeHTTPHandle for web.ServiceHandle
func (make *CCPOEMake) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (make *CCPOEMake) Doc() string {
	return make.document
}
