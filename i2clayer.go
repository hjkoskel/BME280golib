/*
I2C layer there are separate implementations for linux and tinygo
*/
package BME280golib

type I2CDeviceLayer interface {
	WriteReg(address byte, value byte) error
	ReadRegs(address byte, count byte) ([]byte, error)
	Close() error
}
