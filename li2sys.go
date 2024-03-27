//go:build !tinygo

/*
Low level I2C utility for linux golang
*/

package BME280golib

import (
	"fmt"
	"os"
	"syscall"
)

type I2CSys struct {
	f          *os.File
	deviceAddr uint16
}

func CreateI2CSys(f *os.File, deviceAddr uint16) I2CSys {
	return I2CSys{f: f, deviceAddr: deviceAddr}
}

func (p *I2CSys) selectI2CSlave() error {
	//i2c_SLAVE := 0x0703
	_, _, errorcode := syscall.Syscall6(syscall.SYS_IOCTL, p.f.Fd(), 0x0703, uintptr(p.deviceAddr), 0, 0, 0)
	if errorcode != 0 {
		return fmt.Errorf("select I2C slave errcode %v", errorcode)
	}
	return nil
}

func (p *I2CSys) WriteReg(address byte, value byte) error {
	err := p.selectI2CSlave()
	if err != nil {
		return err
	}
	_, err = p.f.Write([]byte{address, value})
	return err
}

func (p *I2CSys) ReadRegs(address byte, count byte) ([]byte, error) {
	selectErr := p.selectI2CSlave()
	result := make([]byte, count)
	if selectErr != nil {
		return result, selectErr
	}

	_, err := p.f.Write([]byte{address})
	if err != nil {
		return result, err
	}
	_, err = p.f.Read(result)
	return result, err
}

func (p *I2CSys) Close() error {
	return p.f.Close()
}
