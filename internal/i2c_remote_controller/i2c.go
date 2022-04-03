package i2c_remote_controller

import (
	"encoding/binary"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

type IRProtocol uint8

const (
	NEC IRProtocol = iota + 1
	ONKYO
)

type Controller struct {
	bus i2c.BusCloser
	dev i2c.Dev
}

func Open(bus string, addr uint16) (*Controller, error) {
	_, err := host.Init()
	if err != nil {
		return nil, err
	}

	b, err := i2creg.Open(bus)
	if err != nil {
		return nil, err
	}

	return &Controller{
		bus: b,
		dev: i2c.Dev{
			Addr: addr,
			Bus:  b,
		},
	}, nil
}

func (c *Controller) Close() error {
	return c.bus.Close()
}

func (c *Controller) Send(proto IRProtocol, addr, cmd uint16) error {
	var msg [5]byte
	msg[0] = byte(proto)
	binary.BigEndian.PutUint16(msg[1:], addr)
	binary.BigEndian.PutUint16(msg[3:], cmd)
	_, err := c.dev.Write(msg[:])
	return err
}
