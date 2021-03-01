package chaincode

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"static/asymmetric"
	gm "static/gm"
	"static/symmetric"
	"testing"

	"../../dcache"
	"../../djson"
	"../../tools"
	"github.com/pkg/errors"
)

func Test_POEWriteData(t *testing.T) {
	var make UserRegisterMake
	make.path = "/poe/writeData"
	make.invoke = true
	make.chaincode = dcache.CCPOENE
	make.function = dcache.MsgForrcpBuild
	make.document = ""
	dcache.PutUserPriv("b1f3da6c6bb18b7a7ef12a3683bd81ca", "LS0tLS1CRUdJTiBFTkNSWVBURUQgUFJJVkFURSBLRVktLS0tLQpNSUg4TUZjR0NTcUdTSWIzRFFFRkRUQktNQ2tHQ1NxR1NJYjNEUUVGRERBY0JBZ0huaVlUWkhLSTBRSUNDQUF3CkRBWUlLb1pJaHZjTkFnY0ZBREFkQmdsZ2hrZ0JaUU1FQVNvRUVDQ2t6dTRXMUVWelJmYS9uTUV6V3JNRWdhQzYKUEV6UnZHTGV5VmhaYkxXT1hzdnB2bm9LNGpOTWJwZXFaWXIwMU5jNzJPd2lXYmpTZC9FVDlIYWFBL0tsc3ltWQpTc2tGRysyTHF2a3hKTzB4UzdoN1dONWdqemwrY0t1bzRxcjZweExkZDdScmdLSkl2N1NXd05oN2Nob0RHSTJRClUyV3o2OFBPMHV6dm9LenowQnoxRmZqVEZNWm15TDJ5QWQ5WEFHYm05S0tEVndXUldjUUJBL05FdUJiUHRSNzUKTWxraGIrRDRMeFNiYmdYeXRYblgKLS0tLS1FTkQgRU5DUllQVEVEIFBSSVZBVEUgS0VZLS0tLS0K")
	dcache.PutUserPrivPwd("b1f3da6c6bb18b7a7ef12a3683bd81ca", "123456")

	body := []byte("{\"address\": \"b1f3da6c6bb18b7a7ef12a3683bd81ca\",\"hash\": \"a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3\",\"data\": \"123\"}")
	//Unmarshal request body
	var ledger Ledger
	err := djson.Unmarshal(body, &ledger)
	if err != nil {
		fmt.Println(err, "Parse input value")
		return
	}

	var hashed []byte
	opt := sha256.New()
	if make.function == dcache.MsgForrcpBuild {
		opt.Write([]byte(ledger.Data))
		sha256 := hex.EncodeToString(opt.Sum(nil))
		if sha256 != ledger.Hash {
			fmt.Println(err, "data hash is not matched.")
			fmt.Println(ledger.Data)
			fmt.Println(sha256)
			fmt.Println(ledger.Hash)
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
		fmt.Println(err, "NewSM4Cipher false.")
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
		fmt.Println(err, "encode IDInfo false.")
		return
	}
	//	fmt.Println(cryptedData)
	opt = sha256.New()
	//	ledger.Data = "123456"
	ledger.Data = tools.Base64Encode(cryptedData)
	opt.Write([]byte(ledger.Data))
	ledger.Hash = hex.EncodeToString(opt.Sum(nil))
	fmt.Println(ledger.Data)
	fmt.Println(ledger.Hash)
	cryptBody, err := djson.Marshal(ledger)
	if err != nil {
		fmt.Println(err, "Marshal false.")
		return
	}

	//create priv with timestamp
	seed, err = asymmetric.NewSeed(string(cryptBody))
	if err != nil {
		fmt.Println(err, "new seed false.")
		return
	}

	//get the login state
	privBase64 := dcache.GetPrivWithAddress(ledger.Address)
	if privBase64 == "" {
		fmt.Println(err, "get priv  false.")
		return
	}

	privPEM, err := tools.Base64Decode([]byte(privBase64))
	if err != nil || len(privPEM) == 0 {
		fmt.Println(err, "Base64Decode privkey false.")
		return
	}
	privPwd := dcache.GetPrivPwdWithAddress(ledger.Address)
	if privPwd == "" {
		fmt.Println(err, "get priv  false.")
		return
	}

	//decode to priv obj
	privKey, err := gm.PEMToPrivateKey(privPEM, []byte(privPwd))
	if err != nil {
		fmt.Println(err, "poe to priv key obj false.")
		return
	}

	opt = sha256.New()
	opt.Write(cryptBody)
	hashed = opt.Sum(nil)
	signed, err := gm.Sign(seed, privKey, hashed)
	if err != nil {
		fmt.Println(err, "sign false.")
		return
	}
	//make arg0
	var signatureInfo SignatureInfo
	signatureInfo.Address = ledger.Address
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
	fmt.Println("arg1:")
	fmt.Println(string(cryptBody))

	fmt.Println("chaincode:")
	fmt.Println(make.chaincode)
	fmt.Println("function:")
	fmt.Println(make.function)

	// objResp, err := fabclient.CallChainCode(
	// 	true, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0), cryptBody})
	// if err != nil {
	// 	return
	// }

	// if objResp.Success == false {
	// 	err = errors.New(objResp.Message)
	// 	return
	// }
	// ret = objResp.Payload
	// id = objResp.TxID
	return
}

func Test_POEVerifyData(t *testing.T) {
	var make UserRegisterMake
	make.path = "/user/verifyData"
	make.invoke = true
	make.chaincode = dcache.CCSIMNI
	make.function = dcache.MsgForrcpVerify
	make.document = ""
	//	ok := true

	dcache.PutUserPriv("b1f3da6c6bb18b7a7ef12a3683bd81ca", "LS0tLS1CRUdJTiBFTkNSWVBURUQgUFJJVkFURSBLRVktLS0tLQpNSUg4TUZjR0NTcUdTSWIzRFFFRkRUQktNQ2tHQ1NxR1NJYjNEUUVGRERBY0JBZ0huaVlUWkhLSTBRSUNDQUF3CkRBWUlLb1pJaHZjTkFnY0ZBREFkQmdsZ2hrZ0JaUU1FQVNvRUVDQ2t6dTRXMUVWelJmYS9uTUV6V3JNRWdhQzYKUEV6UnZHTGV5VmhaYkxXT1hzdnB2bm9LNGpOTWJwZXFaWXIwMU5jNzJPd2lXYmpTZC9FVDlIYWFBL0tsc3ltWQpTc2tGRysyTHF2a3hKTzB4UzdoN1dONWdqemwrY0t1bzRxcjZweExkZDdScmdLSkl2N1NXd05oN2Nob0RHSTJRClUyV3o2OFBPMHV6dm9LenowQnoxRmZqVEZNWm15TDJ5QWQ5WEFHYm05S0tEVndXUldjUUJBL05FdUJiUHRSNzUKTWxraGIrRDRMeFNiYmdYeXRYblgKLS0tLS1FTkQgRU5DUllQVEVEIFBSSVZBVEUgS0VZLS0tLS0K")
	dcache.PutUserPrivPwd("b1f3da6c6bb18b7a7ef12a3683bd81ca", "123456")

	body := []byte("{\"address\": \"b1f3da6c6bb18b7a7ef12a3683bd81ca\",\"txid\": \"a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3\",\"data\": \"123\"}")
	//Unmarshal request body
	var ledger Ledger
	err := djson.Unmarshal(body, &ledger)
	if err != nil {
		fmt.Println(err, "Parse input value")
		return
	}

	var hashed []byte
	opt := sha256.New()
	if make.function == dcache.MsgForrcpBuild {
		opt.Write([]byte(ledger.Data))
		sha256 := hex.EncodeToString(opt.Sum(nil))
		if sha256 != ledger.Hash {
			fmt.Println(err, "data hash is not matched.")
			fmt.Println(ledger.Data)
			fmt.Println(string(opt.Sum(nil)))
			fmt.Println(ledger.Hash)
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
		fmt.Println(err, "NewSM4Cipher false.")
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
		fmt.Println(err, "encode IDInfo false.")
		return
	}
	ledger.Data = tools.Base64Encode(cryptedData)
	opt.Write([]byte(ledger.Data))
	ledger.Hash = hex.EncodeToString(opt.Sum(nil))
	cryptBody, err := djson.Marshal(ledger)
	if err != nil {
		fmt.Println(err, "Marshal false.")
		return
	}

	//create priv with timestamp
	seed, err = asymmetric.NewSeed(string(cryptBody))
	if err != nil {
		fmt.Println(err, "new seed false.")
		return
	}

	//get the login state
	privBase64 := dcache.GetPrivWithAddress(ledger.Address)
	if privBase64 == "" {
		fmt.Println(err, "get priv  false.")
		return
	}

	privPEM, err := tools.Base64Decode([]byte(privBase64))
	if err != nil || len(privPEM) == 0 {
		fmt.Println(err, "Base64Decode privkey false.")
		return
	}
	privPwd := dcache.GetPrivPwdWithAddress(ledger.Address)
	if privPwd == "" {
		fmt.Println(err, "get priv  false.")
		return
	}

	//decode to priv obj
	privKey, err := gm.PEMToPrivateKey(privPEM, []byte(privPwd))
	if err != nil {
		fmt.Println(err, "poe to priv key obj false.")
		return
	}

	opt = sha256.New()
	opt.Write(cryptBody)
	hashed = opt.Sum(nil)
	signed, err := gm.Sign(seed, privKey, hashed)
	if err != nil {
		fmt.Println(err, "sign false.")
		return
	}
	//make arg0
	var signatureInfo SignatureInfo
	signatureInfo.Address = ledger.Address
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
	fmt.Println("arg1:")
	fmt.Println(string(cryptBody))

	fmt.Println("chaincode:")
	fmt.Println(make.chaincode)
	fmt.Println("function:")
	fmt.Println(make.function)

	// objResp, err := fabclient.CallChainCode(
	// 	true, make.invoke, make.chaincode, make.function, [][]byte{[]byte(arg0), cryptBody})
	// if err != nil {
	// 	return
	// }

	// if objResp.Success == false {
	// 	err = errors.New(objResp.Message)
	// 	return
	// }
	// ret = objResp.Payload
	// id = objResp.TxID
	return
}
