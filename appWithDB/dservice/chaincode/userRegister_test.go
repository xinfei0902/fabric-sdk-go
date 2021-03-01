package chaincode

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"static/asymmetric"
	gm "static/gm"
	sm2 "static/gm/sm2"
	"static/symmetric"
	"testing"

	"../../dcache"
	"../../derrors"
	"../../djson"
	"../../tools"
	"github.com/pkg/errors"
)

type ResponseInvokeStruct struct {
	TxID    string
	Success bool
	Message string
	Payload []byte
}

func Test_UserRegister(t *testing.T) {
	var make UserRegisterMake
	make.path = "/user/register"
	make.invoke = true
	make.chaincode = dcache.CCSIMNI
	make.function = dcache.MsgForusrReg
	make.document = ""
	ok := true

	body := []byte("{\"referenceAddress\": \"15454654asdfawefwef\",\"userName\": \"18504596351\",\"userPwd\": \"123456\",\"name\": \"zhangsan\",\"role\": \"角色\",\"uniqueID\": \"afwaef1a5e4f5ae4ff5awe4f5aw4e5f\"}")
	var err error

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
	fmt.Println("arg0:")
	fmt.Println(strArg0)
	strArg1 := string(arg1)
	fmt.Println("arg1:")
	fmt.Println(strArg1)

	fmt.Println("chaincode:")
	fmt.Println(make.chaincode)
	fmt.Println("function:")
	fmt.Println(make.function)
	var objResp ResponseInvokeStruct
	objResp.Success = true

	//call chaincode
	// objResp, err := fabclient.CallChainCode(
	// 	ok, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0), arg1})
	// if err != nil {
	// 	return
	// }
	fmt.Println("response address:", params.Address)
	fmt.Println("response publickey:", params.PublicKey)
	if objResp.Success == false {
		err = errors.New(objResp.Message)
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
	err = ioutil.WriteFile(params.Address, []byte(privBase64), 0644)
	if err != nil {
		err = errors.WithMessage(err, "encrypt priv writeFile false.")
		return
	}
	//	fmt.Println(objResp.TxID)
	return
}

func Test_PrivFindOut(t *testing.T) {
	var make UserRegisterMake
	make.path = "/user/privFindOut"
	make.invoke = false
	make.chaincode = dcache.CCSIMNI
	make.function = dcache.MsgForusrSignVerify
	make.document = ""
	//	ok := true

	body := []byte("{\"referenceAddress\": \"15454646fasdfawefwef\",\"userName\": \"18504596351\",\"userPwd\": \"afwaef1a5e4f5ae4ff5awe4f5aw4e5f\",\"name\": \"zhangsan\",\"role\": \"角色\",\"uniqueID\": \"afwaef1a5e4f5ae4ff5awe4f5aw4e5f\"}")
	var err error

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

	//crypt user priv info use sm4 with user public key
	pubByte, err := gm.PublicKeyToPEM(&priv.PublicKey)
	if err != nil || len(pubByte) == 0 {
		err = errors.WithMessage(err, "PublicKeyToPEM false.")
		return
	}

	params.PublicKey = tools.Base64Encode(pubByte)
	params.Address, err = tools.EncodeSumHashMD5(params.PublicKey)
	if err != nil {
		err = errors.WithMessage(err, "pub to address false.")
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
	fmt.Println("arg0:")
	fmt.Println(strArg0)

	fmt.Println("chaincode:")
	fmt.Println(make.chaincode)
	fmt.Println("function:")
	fmt.Println(make.function)
	var objResp ResponseInvokeStruct
	objResp.Success = false

	//call chaincode
	// objResp, err := fabclient.CallChainCode(
	// 	ok, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0), arg1})
	// if err != nil {
	// 	return
	// }
	fmt.Println("response address:", params.Address)
	fmt.Println("response publickey:", params.PublicKey)
	if objResp.Success == false {
		fmt.Println("response false.")
		return
	}
	//change privkey to byte pem
	privByte, err := gm.PrivateKeyToPEM(priv, []byte(params.UserPwd))
	if err != nil || len(privByte) == 0 {
		err = errors.WithMessage(err, "PrivateKeyToPEM false.")
		return
	}
	privBase64 := tools.Base64Encode(privByte)
	//write priv key to file
	err = ioutil.WriteFile(params.Address, []byte(privBase64), 0644)
	if err != nil {
		err = errors.WithMessage(err, "encrypt priv writeFile false.")
		return
	}
	fmt.Println(objResp.TxID)
	return
}
