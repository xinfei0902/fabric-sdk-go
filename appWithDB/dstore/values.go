package dstore

// KindStore for output service init
type KindStore int

// List of Kind
const (
	NoKind KindStore = iota
	PostgresDBKind
)

func (k KindStore) String() string {
	switch k {
	case PostgresDBKind:
		return "postgres"
	}
	return ""
}

// List of keys in paramaters
const (
	KeyLinkString = "link"
	KeyTables     = "tables"
)
