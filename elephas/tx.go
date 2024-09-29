package elephas

type Tx struct {
}

func NewTransaction() *Tx {
	return &Tx{}
}

func (tx *Tx) Commit() error {
	panic(" not implement")
}

func (tx *Tx) Rollback() error {
	panic(" not implement")
}
