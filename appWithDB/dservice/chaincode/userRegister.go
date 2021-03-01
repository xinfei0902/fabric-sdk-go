package chaincode

import (
	"fmt"
	"net/http"
	"path/filepath"

	"crypto/sha256"
	"static/asymmetric"
	gm "static/gm"
	sm2 "static/gm/sm2"
	"static/symmetric"

	"io/ioutil"

	"../../dconfig"
	"../../derrors"
	"../../djson"
	"../../fabclient"
	"../../tools"
	"../../web"
	"github.com/pkg/errors"
)

// UserRegisterMake for normal handle makes
type UserRegisterMake struct {
	path      string
	invoke    bool
	chaincode string
	function  string
	document  string
}

//Identity 用户信息
type Identity struct {
	ReferenceAddress string `json:"referenceAddress"`
	UserName         string `json:"userName"`
	UserPwd          string `json:"userPwd"`
	Name             string `json:"name"`
	PublicKey        string `json:"publicKey"`
	Role             string `json:"role"`
	Address          string `json:"address"`
	NewPwd           string `json:"newPwd"`
	ConfirmPwd       string `json:"confirmPwd"`
	Status           int    `json:"status"`
	UpdateAddress    string `json:"updateAddress"`
	PreTxid          string `json:"preTxid"`
	ResetPwd         string `json:"resetPwd"`
	ResetConfirmPwd  string `json:"resetConfirmPwd"`
	UniqueID         string `json:"uniqueID"`
	ExpireTime       string `json:"expireTime"`
}

// NewUserRegisterMake create one make
func NewUserRegisterMake(path string, invoke bool, chaincode string, function string, document string) *UserRegisterMake {
	return &UserRegisterMake{path, invoke, chaincode, function, document}
}

// Path for web.ServiceHandle
func (make *UserRegisterMake) Path() string {
	return make.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (make *UserRegisterMake) MakeChainCodeHandle() web.ChainCodeHandle {

	rootPath := dconfig.GetStringByKey("walletpath")

	return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
		if len(body) == 0 {
			err = derrors.ErrorEmptyValue
			return
		}
		//Unmarshal request body
		var params Identity
		err = djson.Unmarshal(body, &params)
		if err != nil {
			err = errors.WithMessage(err, "Parse input value")
			return
		}
		if len(params.UserPwd) == 0 || len(params.UniqueID) == 0 {
			err = errors.WithMessage(err, "userPwd and uniqueID is not null")
			return
		}
		privEncodePwd := params.UserPwd

		//create priv with pwd and ID
		seed, err := asymmetric.NewSeed(params.UserName + params.Name + params.UniqueID)
		if err != nil {
			err = errors.WithMessage(err, "new seed false.")
			return
		}
		//create priv
		priv, err := sm2.GenerateKey(seed)
		if err != nil {
			err = errors.WithMessage(err, "GenerateKey false.")
			return
		}
		//sign with register body
		//		context := []byte(body)
		opt := sha256.New()
		opt.Write(body)
		hashed := opt.Sum(nil)
		signed, err := gm.Sign(seed, priv, hashed)
		if err != nil {
			err = errors.WithMessage(err, "sign false.")
			return
		}

		_, ok := args["async"]

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
		cryptedID, err := symmetric.EncodeByBlock(block, nil, []byte(params.UniqueID))
		if err != nil {
			err = errors.WithMessage(err, "encode IDInfo false.")
			return
		}
		cryptedName, err := symmetric.EncodeByBlock(block, nil, []byte(params.Name))
		if err != nil {
			err = errors.WithMessage(err, "encode IDInfo false.")
			return
		}
		cryptedUserName, err := symmetric.EncodeByBlock(block, nil, []byte(params.UserName))
		if err != nil {
			err = errors.WithMessage(err, "encode IDInfo false.")
			return
		}

		params.UserPwd = ""
		params.Name = tools.Base64Encode(cryptedName)
		params.UserName = tools.Base64Encode(cryptedUserName)
		params.UniqueID = tools.Base64Encode(cryptedID)
		params.PublicKey = tools.Base64Encode(pubByte)
		params.Address, err = tools.EncodeSumHashMD5(params.PublicKey)
		if err != nil {
			err = errors.WithMessage(err, "pub to address false.")
			return
		}
		arg1, err := djson.Marshal(params)
		if err != nil {
			err = errors.WithMessage(err, "Marshal arg1 false.")
			return
		}

		//make arg0
		var signatureInfo SignatureInfo
		signatureInfo.Address = params.Address
		signatureInfo.RandomCode = hashed
		signatureInfo.Signature = signed
		arg0, err := djson.Marshal(&signatureInfo)
		if err != nil {
			err = errors.WithMessage(err, "signatureInfo marsha1 false.")
			return
		}

		strArg0 := string(arg0)
		fmt.Println(strArg0)
		strArg1 := string(arg1)
		fmt.Println(strArg1)

		//call chaincode
		objResp, err := fabclient.CallChainCode(
			ok, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0), arg1})
		if err != nil {
			return "", nil, err
		}
		ret = fmt.Sprintf("{\"address\":\"%s\",\"publicKey\":\"%s\"}",
			params.Address, params.PublicKey)
		if objResp.Success == false {
			err = errors.WithMessage(err, "call chaincode false.")
			return
		}
		//change privkey to byte pem
		privByte, err := gm.PrivateKeyToPEM(priv, []byte(privEncodePwd))
		if err != nil || len(privByte) == 0 {
			err = errors.WithMessage(err, "PrivateKeyToPEM false.")
			return
		}
		privBase64 := tools.Base64Encode(privByte)
		//write priv key to file
		err = ioutil.WriteFile(filepath.Join(rootPath, params.Address), []byte(privBase64), 0644)
		if err != nil {
			err = errors.WithMessage(err, "encrypt priv writeFile false.")
			return
		}
		id = objResp.TxID
		return
	}
}

// MakeHTTPHandle for web.ServiceHandle
func (make *UserRegisterMake) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (make *UserRegisterMake) Doc() string {
	return make.document
}
