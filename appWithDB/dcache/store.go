package dcache

import (
	memdb "github.com/hashicorp/go-memdb"
)

var globalOpt *memdb.MemDB

// Const Table name & index key
const (
	TablePeerKey   = "peers"
	AddressPeerKey = "id"

	TableBlockKey  = "block"
	HeightBlockKey = "id"
	HashBlockKey   = "hash"
)

func peerCache() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: TablePeerKey,
		Indexes: map[string]*memdb.IndexSchema{
			AddressPeerKey: &memdb.IndexSchema{
				Name:    AddressPeerKey,
				Unique:  true,
				Indexer: &memdb.StringFieldIndex{Field: "PeerAddress"},
			},
		},
	}
}

func blockCache() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: TableBlockKey,
		Indexes: map[string]*memdb.IndexSchema{
			HeightBlockKey: &memdb.IndexSchema{
				Name:    HeightBlockKey,
				Unique:  true,
				Indexer: &memdb.UintFieldIndex{Field: "Number"},
			},
			HashBlockKey: &memdb.IndexSchema{
				Name:    HashBlockKey,
				Unique:  true,
				Indexer: &BlockDataHashIndex{},
			},
		},
	}
}
