package elephas

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net/url"
)

const name = "elephas"

type Driver struct {
	connector *Connector
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

func NewDriver() *Driver {
	return &Driver{}
}

func init() {
	sql.Register(name, NewDriver())
}
