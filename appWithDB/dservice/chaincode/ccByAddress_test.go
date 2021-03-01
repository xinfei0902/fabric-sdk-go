package chaincode

import (
	"crypto/sha256"
	"fmt"
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

// UserInfo for chaincode userquery
func UserInfo(priv *sm2.PrivateKey, payload []byte) (ret []byte, err error) {
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
func POEGetData(priv *sm2.PrivateKey, payload []byte) (ret []byte, err error) {
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

func Test_CCByAddress(t *testing.T) {
	var make UserRegisterMake
	make.path = "/asset/writeData"
	make.invoke = true
	make.chaincode = dcache.CCATCNT
	make.function = dcache.MsgForrcpInfo
	make.document = ""
	//	ok := true

	dcache.PutUserPriv("b1f3da6c6bb18b7a7ef12a3683bd81ca", "LS0tLS1CRUdJTiBFTkNSWVBURUQgUFJJVkFURSBLRVktLS0tLQpNSUg4TUZjR0NTcUdTSWIzRFFFRkRUQktNQ2tHQ1NxR1NJYjNEUUVGRERBY0JBZ0huaVlUWkhLSTBRSUNDQUF3CkRBWUlLb1pJaHZjTkFnY0ZBREFkQmdsZ2hrZ0JaUU1FQVNvRUVDQ2t6dTRXMUVWelJmYS9uTUV6V3JNRWdhQzYKUEV6UnZHTGV5VmhaYkxXT1hzdnB2bm9LNGpOTWJwZXFaWXIwMU5jNzJPd2lXYmpTZC9FVDlIYWFBL0tsc3ltWQpTc2tGRysyTHF2a3hKTzB4UzdoN1dONWdqemwrY0t1bzRxcjZweExkZDdScmdLSkl2N1NXd05oN2Nob0RHSTJRClUyV3o2OFBPMHV6dm9LenowQnoxRmZqVEZNWm15TDJ5QWQ5WEFHYm05S0tEVndXUldjUUJBL05FdUJiUHRSNzUKTWxraGIrRDRMeFNiYmdYeXRYblgKLS0tLS1FTkQgRU5DUllQVEVEIFBSSVZBVEUgS0VZLS0tLS0K")
	dcache.PutUserPrivPwd("b1f3da6c6bb18b7a7ef12a3683bd81ca", "123456")

	body := []byte("{\"address\": \"b1f3da6c6bb18b7a7ef12a3683bd81ca\",\"symbol\": \"a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3\",\"name\": \"aaaa\",\"amount\": 123,\"rule\": \"123\",\"issuer\": \"123\",\"hash\": \"123\",\"metas\": \"123\"}")
	//Unmarshal request body
	var signatureInfo SignatureInfo
	err := djson.Unmarshal(body, &signatureInfo)
	if err != nil {
		fmt.Println(err, "Parse input value")
		return
	}
	privBase64 := dcache.GetPrivWithAddress(signatureInfo.Address)
	if privBase64 == "" {
		fmt.Println("get priv false")
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

	strArg0 := string(arg0)
	fmt.Println("arg0:")
	fmt.Println(strArg0)
	fmt.Println("arg1:")
	fmt.Println(string(body))

	fmt.Println("chaincode:")
	fmt.Println(make.chaincode)
	fmt.Println("function:")
	fmt.Println(make.function)

	// _, ok := args["async"]
	// objResp, err := fabclient.CallChainCode(
	// 	ok, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0), body})
	// if err != nil {
	// 	return "", nil, err
	// }

	// if objResp.Success == false {
	// 	err = errors.New(objResp.Message)
	// 	return
	// }
	var ret []byte
	var objResp ResponseInvokeStruct

	switch make.function {
	case dcache.MsgForusrInfo:
		objResp.Payload = []byte("456")
		ret, err = UserInfo(privKey, objResp.Payload)
		if err != nil {
			fmt.Println(err, "user info query false.")
			return
		}
	case dcache.MsgForrcpInfo:
		objResp.Payload = []byte("789")
		fmt.Println("user info query false.")
		ret, err = POEGetData(privKey, objResp.Payload)
		if err != nil {
			err = errors.WithMessage(err, "user info query false.")
			return
		}
	default:
		ret = []byte("123")
	}
	fmt.Println(string(ret))
	//	id = objResp.TxID
	return
}
