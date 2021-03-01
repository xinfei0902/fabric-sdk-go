package dcmd

import (
	"fmt"

	"../dconfig"
	"../web"
)

func addWebFlag() error {
	GlobalDefaultAddress := ":7000"

	err := dconfig.Register("", "cert", "", "Cert File name for TLS")
	if err != nil {
		return err
	}
	err = dconfig.Register("", "key", "", "Key File name for TLS")
	err = dconfig.Register("a", "address", GlobalDefaultAddress, "Bind Service on this Address. Default: "+GlobalDefaultAddress)
	return err
}

// StartWeb and hold calling thread
func StartWeb() (err error) {
	cert := dconfig.GetStringByKey("cert")
	key := dconfig.GetStringByKey("key")
	address := dconfig.GetStringByKey("address")
	if len(cert) > 0 && len(key) > 0 {
		fmt.Println("Start TLS service on", `"`+address+`"`)
		err = web.StartTLSService(address, cert, key)
	} else {
		fmt.Println("Start service on", `"`+address+`"`)
		err = web.StartService(address)
	}
	return
}
