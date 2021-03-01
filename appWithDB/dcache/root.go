package dcache

import memdb "github.com/hashicorp/go-memdb"

func init() {

	peerSchema := &memdb.DBSchema{}
	peerSchema.Tables = make(map[string]*memdb.TableSchema)

	peerSchema.Tables[TablePeerKey] = peerCache()
	peerSchema.Tables[TableBlockKey] = blockCache()

	var err error
	globalOpt, err = memdb.NewMemDB(peerSchema)
	if err != nil {
		panic(err)
	}
}
