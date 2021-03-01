package dcmd

import (
	"static/mail"

	"../dconfig"
)

func addMailFlags() (err error) {
	dconfig.Register("", "mailidentify", "", "Mail identify")
	dconfig.Register("", "mailtarget", "", "Mail sendto")
	dconfig.Register("", "mailuser", "", "Mail username")
	dconfig.Register("", "mailpassword", "", "Mail passwrod")
	dconfig.Register("", "mailhost", "", "Mail service host")
	dconfig.Register("", "mailaddress", "", "Mail service IP:PORT")
	return nil
}

func startMail() (err error) {
	identify := dconfig.GetStringByKey("mailidentify")
	user := dconfig.GetStringByKey("mailuser")
	pwd := dconfig.GetStringByKey("mailpassword")
	host := dconfig.GetStringByKey("mailhost")
	addr := dconfig.GetStringByKey("mailaddress")

	return mail.Init(identify, user, pwd, host, addr)
}
