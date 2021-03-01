package dcmd

import (
	"../dconfig"
	"../dstore"
)

func addPostgresFlags() error {
	return dconfig.Register("l", "link", "", "DB link strings")
}

func getPostgresFlags() string {
	return dconfig.GetStringByKey("link")
}

// StartDB by link string
func StartDB(link string) error {
	return dstore.Init(dstore.PostgresDBKind, map[string]interface{}{
		dstore.KeyLinkString: link,
	})
}
