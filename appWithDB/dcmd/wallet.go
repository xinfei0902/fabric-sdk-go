package dcmd

import (
	"os"

	"../dconfig"
)

func addWalletFlags() (err error) {
	dconfig.Register("", "walletpath", "", "Path of Wallet keys")
	dconfig.Register("", "crypto", "", "Crypto names")
	return nil
}

func startWallet() (err error) {
	wp := dconfig.GetStringByKey("walletpath")
	if len(wp) > 0 {
		err = os.MkdirAll(wp, 0755)
	}
	return err
}
