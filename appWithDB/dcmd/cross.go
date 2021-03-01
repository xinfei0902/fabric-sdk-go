package dcmd

import (
	"../dconfig"

	"restapi/m"
)

func addCrossFlags() (err error) {
	dconfig.Register("", "btip", "127.0.0.1", "Bitcoin service IP")
	dconfig.Register("", "btport", 8346, "Bitcoin service Port")
	dconfig.Register("", "btuser", "", "Bitcoin service Wallet user")
	dconfig.Register("", "btpwd", "", "Bitcoin service Wallet password")
	return nil
}

func startCrossSDK() error {
	ip := dconfig.GetStringByKey("btip")
	port := dconfig.GetIntByKey("btport")
	usr := dconfig.GetStringByKey("btuser")
	pwd := dconfig.GetStringByKey("btpwd")

	return m.Init(usr, pwd, ip, port)
}
