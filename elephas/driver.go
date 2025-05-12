package elephas

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"net/url"
	"strings"
)

const name = "elephas"

type Driver struct {
	connector *Connector
}

func NewDriver() *Driver {
	return &Driver{}
}

func init() {
	sql.Register(name, NewDriver())
}

func (d *Driver) Open(name string) (driver.Conn, error) {
	connector, err := d.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return connector.Connect(context.TODO())
}

func (d *Driver) OpenConnector(dsn string) (driver.Connector, error) {
	cfg, err := d.parse(dsn)
	if err != nil {
		return nil, err
	}
	return NewConnector(cfg), nil
}

func (d *Driver) parse(dsn string) (*Config, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	passw, _ := u.User.Password()
	return &Config{
		Network:  "tcp",
		Addr:     u.Host,
		User:     u.User.Username(),
		Password: passw,
		Database: u.Path[1:],
	}, nil
}

func ReadCommandComplete(r *Reader) ([]string, error) {
	_, err := r.Read4Bytes()
	if err != nil {
		return nil, err
	}
	tag, err := r.ReadString(0)
	if err != nil {
		return nil, err
	}

	return strings.Split(tag[:len(tag)-1], " "), io.EOF
}
