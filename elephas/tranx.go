package elephas

type tranx struct {
}

func (tx tranx) Commit() error {
	panic(" not implement")
}
func (tx tranx) Rollback() error {
	panic(" not implement")
}
