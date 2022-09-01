/*
Low level I2C utility functions
*/
package BME280golib

import (
	"fmt"
	"os"
	"syscall"
)

func selectI2CSlave(f *os.File, address byte) error {
	//i2c_SLAVE := 0x0703
	_, _, errorcode := syscall.Syscall6(syscall.SYS_IOCTL, f.Fd(), 0x0703, uintptr(address), 0, 0, 0)
	if errorcode != 0 {
		return fmt.Errorf("Select I2C slave errcode %v", errorcode)
	}
	return nil
}

func (p *BME280) writeReg(address byte, value byte) error {
	err := selectI2CSlave(p.i2cHandle, p.address)
	if err != nil {
		return err
	}
	_, err = p.i2cHandle.Write([]byte{address, value})
	return err
}

func (p *BME280) readRegs(address byte, count byte) ([]byte, error) {
	selectErr := selectI2CSlave(p.i2cHandle, p.address)
	result := make([]byte, count)
	if selectErr != nil {
		return result, selectErr
	}

	_, err := p.i2cHandle.Write([]byte{address})
	if err != nil {
		return result, err
	}
	_, err = p.i2cHandle.Read(result)
	return result, err
}
