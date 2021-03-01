package chaincode

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"static/asymmetric"
	gm "static/gm"
	"testing"

	"../../dcache"
	"../../djson"
	"../../tools"
	"github.com/pkg/errors"
)

func Test_UserLogin(t *testing.T) {
	var make UserRegisterMake
	make.path = "/user/register"
	make.invoke = true
	make.chaincode = dcache.CCSIMNI
	make.function = dcache.MsgForusrReg
	make.document = ""
	//	ok := true

	body := []byte("{\"address\": \"b1f3da6c6bb18b7a7ef12a3683bd81ca\",\"userPwd\": \"123456\"}")
	var params Identity
	err := djson.Unmarshal(body, &params)
	if err != nil {
		fmt.Println(err, "Parse input value")
		return
	}
	if len(params.UserPwd) == 0 || len(params.Address) == 0 {
		fmt.Println(err, "userPwd and Address is not null")
		return
	}

	//read priv file
	privBuff, err := ioutil.ReadFile(params.Address)
	if err != nil || len(privBuff) == 0 {
		fmt.Println(err, "ReadFile priv file false or priv is nil.")
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
		fmt.Println(err, "poe to priv key obj false.")
		return
	}
	// int64timeUnixNano := time.Now().UnixNano()
	// strtimeUnixNano := strconv.FormatInt(int64timeUnixNano, 10)
	// //create priv with timestamp
	seed, err := asymmetric.NewSeed(string(body))
	if err != nil {
		fmt.Println(err, "new seed false.")
		return
	}
	//sign with timestamp
	//		context := []byte(body)
	opt := sha256.New()
	opt.Write(body)
	hashed := opt.Sum(nil)
	signed, err := gm.Sign(seed, privKey, hashed)
	if err != nil {
		fmt.Println(err, "sign false.")
		return
	}
	//make arg1
	var signatureInfo SignatureInfo
	signatureInfo.Address = params.Address
	signatureInfo.RandomCode = hashed
	signatureInfo.Signature = signed
	arg0, err := djson.Marshal(&signatureInfo)
	if err != nil {
		fmt.Println(err, "signatureInfo marsha1 false.")
		return
	}

	strArg0 := string(arg0)
	fmt.Println("arg0:")
	fmt.Println(strArg0)

	fmt.Println("chaincode:")
	fmt.Println(make.chaincode)
	fmt.Println("function:")
	fmt.Println(make.function)

	var objResp ResponseInvokeStruct
	objResp.Success = true

	//	_, ok := args["async"]

	// objResp, err := fabclient.CallChainCode(
	// 	ok, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0)})
	// if err != nil {
	// 	return "", nil, err
	// }

	if objResp.Success == false {
		err = errors.New(objResp.Message)
		return
	}
	//write priv to cache，k = address   v = priv;
	//cache cleared when userLogout or app exit;
	fmt.Println(string(privBuff))
	dcache.PutUserPriv(params.Address, string(privBuff))
	dcache.PutUserPrivPwd(params.Address, params.UserPwd)

	// id = objResp.TxID
	// ret = objResp.Payload
	return
}
