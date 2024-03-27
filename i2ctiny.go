//go:build tinygo

package BME280golib

import "tinygo.org/x/drivers"

type I2CTiny struct {
	dev        drivers.I2C
	deviceAddr uint16
}

func CreateI2CTiny(dev drivers.I2C, deviceAddr uint16) I2CTiny {
	return I2CTiny{dev: dev, deviceAddr: deviceAddr}
}

func (p *I2CTiny) WriteReg(address byte, value byte) error {
	return p.dev.Tx(p.deviceAddr, []byte{address, value}, nil)
}

func (p *I2CTiny) ReadRegs(address byte, count byte) ([]byte, error) {
	output := make([]byte, count)
	txErr := p.dev.Tx(p.deviceAddr, []byte{address}, output)
	return output, txErr
}

func (p *I2CTiny) Close() error {
	return nil
}
