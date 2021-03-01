package chaincode

import (
	"sync"

	"../../dcache"
	"../../dstore"
	"../../web"
)

// Todo here
// Add more push actions
var globalStaticOption []dstore.MakePusher
var onceDB sync.Once

// PushOption for working option
func PushOption() []dstore.PushOption {
	// ToDo here
	// resigterTables in convert
	onceDB.Do(func() {
		globalStaticOption = make([]dstore.MakePusher, 0, 16)

		// Add here to stored datas to DB
		globalStaticOption = append(globalStaticOption, NewUserTrace)
	})

	if len(globalStaticOption) == 0 {
		return nil
	}

	ret := make([]dstore.PushOption, len(globalStaticOption))
	for i, m := range globalStaticOption {
		ret[i] = m()
	}

	return ret
}

// Todo here
var globalCCAPI []web.ServiceHandle
var onceWeb sync.Once

// WebAPI for Web API
func WebAPI() []web.ServiceHandle {
	onceWeb.Do(func() {
		globalCCAPI = make([]web.ServiceHandle, 0, 64)
		globalCCAPI = append(globalCCAPI, webAPIForCross()...)
		globalCCAPI = append(globalCCAPI, webAPIForOldAPI()...)
		globalCCAPI = append(globalCCAPI, webAPIV1call()...)
		// Web API add here
		//globalCCAPI = []web.ServiceHandle{}
	})

	return globalCCAPI
}
func webAPIV1call() []web.ServiceHandle {
	// A call BaaS start status
	// A call BaaS send A money done
	// B call BaaS send B money done
	// A call BaaS get B money
	// check status
	return []web.ServiceHandle{
		&ToolsCC{
			"/v1/call",
			``,
		},
	}
}

// WebAPIForDebug for Web API
func WebAPIForDebug() []web.ServiceHandle {
	return []web.ServiceHandle{
		NewSingleKeyMake("/debug/put", true, "example_cc", "put", "key", ``),
		NewSingleKeyMake("/debug/get", false, "example_cc", "get", "key", ``),
		NewSingleKeyMake("/debug/search", false, "example_cc", "search", "key", ``),
		NewSingleKeyMake("/debug/del", true, "example_cc", "del", "key", ``),
		NewSingleKeyMake("/debug/oput", true, "example_cc", "put", "key", ``),
		NewSingleKeyMake("/debug/oget", false, "example_cc", "get", "key", ``),

		NewSingleKeyMake("/debug/cput", true, "example_cc", "cput", "key", ``),
		NewSingleKeyMake("/debug/cget", false, "example_cc", "cget", "key", ``),
		NewSingleKeyMake("/debug/csearch", false, "example_cc", "csearch", "key", ``),
	}
}

func webAPIForOldAPI() []web.ServiceHandle {
	return []web.ServiceHandle{
		NewUserRegisterMake("/user/register", true, dcache.CCSIMNI, dcache.MsgForusrReg, ""),
		NewUserRegisterMake("/user/privFindOut", true, dcache.CCSIMNI, dcache.MsgForusrSignVerify, ""),
		NewUserLoginMake("/user/login", true, dcache.CCSIMNI, dcache.MsgForusrLogin, ""),
		NewCCByAddressMake("/user/logout", true, dcache.CCSIMNI, dcache.MsgForusrLogout, ""),
		NewCCByAddressMake("/user/logoff", true, dcache.CCSIMNI, dcache.MsgForusrLogoff, ""),
		NewCCByAddressMake("/user/query", false, dcache.CCSIMNI, dcache.MsgForusrInfo, ""),
		NewCCByAddressMake("/user/changepwd", true, dcache.CCSIMNI, dcache.MsgForpwdChange, ""),
		NewCCByAddressMake("/user/resetPwd", true, dcache.CCSIMNI, dcache.MsgForpwdReset, ""),
		NewCCByAddressMake("/user/update", true, dcache.CCSIMNI, dcache.MsgForusrUpdate, ""),
		NewCCByAddressMake("/user/review", true, dcache.CCSIMNI, dcache.MsgForusrReview, ""),
		//		NewCCByAddressMake("/user/statistics", true,  dcache.CCSIMNI, "usrAudit", ""),
		//		NewCCByAddressMake("/user/list", false,  dcache.CCSIMNI, "usrList", ""),
		NewCCByAddressMake("/user/updateHistory", false, dcache.CCSIMNI, dcache.MsgForusrHistory, ""),
		NewCCByAddressMake("/user/input", true, dcache.CCSIMNI, dcache.MsgForusrInput, ""),

		NewCCByAddressMake("/asset/account/import", true, dcache.CCATCNT, dcache.MsgForastImport, ""),
		NewCCByAddressMake("/asset/account/export", true, dcache.CCATCNT, dcache.MsgForastExport, ""),
		NewCCByAddressMake("/asset/account/changeStatus", true, dcache.CCATCNT, dcache.MsgForastChangeStatus, ""),
		NewCCByAddressMake("/asset/account/updateHistory", false, dcache.CCATCNT, dcache.MsgForastUpdateHistory, ""),
		NewCCByAddressMake("/asset/account/info", false, dcache.CCATCNT, dcache.MsgForastAccountInfo, ""),
		NewCCByAddressMake("/asset/account/logoff", true, dcache.CCATCNT, dcache.MsgForastAccountLogoff, ""),
		NewCCByAddressMake("/asset/writeData", true, dcache.CCATCNT, dcache.MsgForastPublish, ""),
		NewCCByAddressMake("/asset/issue", true, dcache.CCATCNT, dcache.MsgForastIssue, ""),
		NewCCByAddressMake("/asset/transaction", true, dcache.CCATCNT, dcache.MsgForastTrade, ""),
		NewCCPublicOperMake("/asset/list", false, dcache.CCATCNT, dcache.MsgForastList, ""),
		NewCCByAddressMake("/asset/detail", false, dcache.CCATCNT, dcache.MsgForastInfo, ""),
		NewCCByAddressMake("/asset/balance", false, dcache.CCATCNT, dcache.MsgForastAmount, ""),
		NewCCByAddressMake("/asset/status", false, dcache.CCATCNT, dcache.MsgForastStatus, ""),
		NewCCForDebugMake("/asset/transaction/changeValue", true, dcache.CCATCNT, dcache.MsgForastTestChangeValue, ""),

		NewCCPOEMake("/poe/writeData", true, dcache.CCPOENE, dcache.MsgForrcpBuild, ""),
		NewCCByAddressMake("/poe/getData", false, dcache.CCPOENE, dcache.MsgForrcpInfo, ""),
		NewCCPOEMake("/poe/verifyData", false, dcache.CCPOENE, dcache.MsgForrcpVerify, ""),
		NewCCByAddressMake("/poe/authorization", true, dcache.CCPOENE, dcache.MsgForrcpAuth, ""),

		//		NewOldFashionMake("/uba/getDetailByID", false, "UBANO", "actInfo", "", ``),
		//		NewOldFashionMake("/uba/recordsQuery", false, "UBANO", "actTrace", "", ``),
		//		NewOldFashionMake("/uba/statistics", false, "UBANO", "actAudit", "", ``),
	}
}

func webAPIForCross() []web.ServiceHandle {
	// A call BaaS start status
	// A call BaaS send A money done
	// B call BaaS send B money done
	// A call BaaS get B money
	// check status
	return []web.ServiceHandle{
		NewCrossChain("/cross/v0/issue", "cross", "issue", ``),
		NewCrossChain("/cross/v0/increase", "cross", "increase", ``),
		NewCrossChain("/cross/v0/fetch", "cross", "fetch", ``),
		NewCrossChainV1("/cross/v1/transcation", "cross", ``),
		NewCrossChainV2("/cross/v2/start", "cross", "start", ``),
		NewCrossChainV2("/cross/v2/cancel", "cross", "cancel", ``),
		NewCrossChainV2("/cross/v2/search", "cross", "search", ``),
		NewCrossChainV2("/cross/v2/complete", "cross", "complete", ``),
	}
}
