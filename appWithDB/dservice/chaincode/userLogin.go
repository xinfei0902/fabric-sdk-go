package chaincode

import (
	"fmt"
	"net/http"
	"path/filepath"

	"crypto/sha256"
	"static/asymmetric"
	gm "static/gm"

	"io/ioutil"

	"../../dcache"
	"../../dconfig"
	"../../derrors"
	"../../djson"
	"../../fabclient"
	"../../tools"
	"../../web"
	"github.com/pkg/errors"
)

// UserLoginMake for normal handle makes
type UserLoginMake struct {
	path      string
	invoke    bool
	chaincode string
	function  string
	document  string
}

// NewUserLoginMake create one make
func NewUserLoginMake(path string, invoke bool, chaincode string, function string, document string) *UserLoginMake {
	return &UserLoginMake{path, invoke, chaincode, function, document}
}

// Path for web.ServiceHandle
func (make *UserLoginMake) Path() string {
	return make.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (make *UserLoginMake) MakeChainCodeHandle() web.ChainCodeHandle {

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
		if len(params.UserPwd) == 0 || len(params.Address) == 0 {
			err = errors.WithMessage(err, "userPwd and Address is not null")
			return
		}

		//read priv file
		privBuff, err := ioutil.ReadFile(filepath.Join(rootPath, params.Address))
		if err != nil || len(privBuff) == 0 {
			err = errors.WithMessage(err, "ReadFile priv file false or priv is nil.")
			return
		}

		privBase64, err := tools.Base64Decode(privBuff)
		if err != nil || len(privBase64) == 0 {
			fmt.Println(err, "Base64Decode privkey false.")
			return
		}

		//decode to priv obj
		privKey, err := gm.PEMToPrivateKey(privBase64, []byte(params.UserPwd))
		if err != nil {
			err = errors.WithMessage(err, "poe to priv key obj false.")
			return
		}

		// //create priv with timestamp
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
		//make arg1
		var signatureInfo SignatureInfo
		signatureInfo.Address = params.Address
		signatureInfo.RandomCode = hashed
		signatureInfo.Signature = signed
		arg0, err := djson.Marshal(&signatureInfo)
		if err != nil {
			err = errors.WithMessage(err, "signatureInfo marsha1 false.")
			return
		}

		_, ok := args["async"]

		objResp, err := fabclient.CallChainCode(
			ok, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0)})
		if err != nil {
			return "", nil, err
		}

		if objResp.Success == false {
			err = errors.New(objResp.Message)
			return
		}
		//write priv to cache，k = address   v = priv;
		//cache cleared when userLogout or app exit;
		dcache.PutUserPriv(params.Address, string(privBuff))
		dcache.PutUserPrivPwd(params.Address, params.UserPwd)

		id = objResp.TxID
		ret = objResp.Payload
		return
	}
}

// MakeHTTPHandle for web.ServiceHandle
func (make *UserLoginMake) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (make *UserLoginMake) Doc() string {
	return make.document
}
