package dcache

var (
	// For cache wallet priv
	userPriv = make(map[string]string)

	//POEDataEncodePwd write data pwd
	POEDataEncodePwd = "scos@sinochain.com"

	//POEDataEncodePwd write data pwd
	userPrivPwd = make(map[string]string)
)

//PutUserPriv write priv to cache
func PutUserPriv(address string, priv string) {
	userPriv[address] = priv
}

//GetPrivWithAddress get priv from cache
func GetPrivWithAddress(address string) string {
	return userPriv[address]
}

//PutUserPrivPwd write privPwd to cache
func PutUserPrivPwd(address string, privPwd string) {
	userPrivPwd[address] = privPwd
}

//GetPrivPwdWithAddress get privPwd from cache
func GetPrivPwdWithAddress(address string) string {
	return userPrivPwd[address]
}

//RemoveCacheData get privPwd from cache
func RemoveCacheData(address string) {
	delete(userPriv, address)
	delete(userPrivPwd, address)
}
