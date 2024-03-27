# BME280golib
Golang library for BME280 I2C temperature - humidity - pressure sensor

Please check ./sensortest subdirectory as example how to use this library


This library has been updated for tinygo compatibility

Actual device is defined as BME280Device 
``` go 
type BME280Device interface {
	Close() error
	Configure(config BME280Config) error
	Read() (HumTempPressureMeas, error)
	SoftReset() error
	GetCalibration() (CalibrationRegs, error) //Gets latest values
}
```


At the moment only I2C interface is **BME280I2C** implementing this interface. Later there might be SPI version of sensor if needed.

The **BME280I2C** is created by function

``` go 
func CreateBME280I2C(layer I2CDeviceLayer) (BME280I2C, error) {
```

There are two implementations of **I2CDeviceLayer**. 
For tinygo **I2CTiny** and for linux **I2CSys**

``` go
func CreateI2CTiny(dev drivers.I2C, deviceAddr uint16) I2CTiny {
func CreateI2CSys(f *os.File, address byte) I2CSys {
```


