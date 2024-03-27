/*
Library for reading BME280 from i2c device file
*/
package BME280golib

const (
	ID_EXPECTED = 0x60
)

const (
	// Register
	REGISTER_ID        byte = 0xD0
	REGISTER_RESET     byte = 0xE0
	REGISTER_CTRL_HUM  byte = 0xF2
	REGISTER_STATUS    byte = 0xF3 //Bit0=update bit3=measuring
	REGISTER_CTRL_MEAS byte = 0xF4
	REGISTER_CONFIG    byte = 0xF5

	REGISTER_DATA byte = 0xF7
	/*
		REGISTER_PRESS_MSB  = 0xF7 //RawMeas
		REGISTER_PRESS_LSB  = 0xF8
		REGISTER_PRESS_XLSB = 0xF9
		REGISTER_TEMP_MSB   = 0xFA
		REGISTER_TEMP_LSB   = 0xFB
		REGISTER_TEMP_XLSB  = 0xFC
		REGISTER_HUM_MSB    = 0xFD
		REGISTER_HUM_LSB    = 0xFE
	*/

	REGISTER_CALIB00 byte = 0x88
	REGISTER_CALIB26 byte = 0xE1
)

// BME280Device inteface, for faking sensor if needed
type BME280Device interface {
	Close() error
	Configure(config BME280Config) error
	Read() (HumTempPressureMeas, error)
	SoftReset() error
	GetCalibration() (CalibrationRegs, error) //Gets latest values
}
