/*
Library for reading BME280 from i2c device file
*/
package BME280golib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
)

const (
	// Register
	REGISTER_ID        = 0xD0
	REGISTER_RESET     = 0xE0
	REGISTER_CTRL_HUM  = 0xF2
	REGISTER_STATUS    = 0xF3 //Bit0=update bit3=measuring
	REGISTER_CTRL_MEAS = 0xF4
	REGISTER_CONFIG    = 0xF5

	REGISTER_DATA = 0xF7
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

	REGISTER_CALIB00 = 0x88
	REGISTER_CALIB26 = 0xE1
)

type BME280 struct {
	I2CHandle *os.File //Set this at startup
	Calib1    CalibrationRegs1
	Calib2    CalibrationRegs2
	Address   byte
}

type HumTempPressureMeas struct {
	Temperature float64
	Rh          float64
	Pressure    float64
}

type RawMeas struct {
	Pmsb  byte
	Plsb  byte
	Pxlsb byte

	Tmsb  byte
	Tlsb  byte
	Txlsb byte

	Hmsb byte
	Hlsb byte
}

const (
	OVRSAMPLE_NO = 0
	OVRSAMPLE_1  = 1
	OVRSAMPLE_2  = 2
	OVRSAMPLE_4  = 3
	OVRSAMPLE_16 = 4
)

type Oversample byte

const (
	MODE_SLEEP  = 0
	MODE_FORCED = 1
	MODE_NORMAL = 3
)

type DeviceMode byte

const (
	STANDBYDURATION_0_5  = 0
	STANDBYDURATION_62_5 = 1
	STANDBYDURATION_125  = 2
	STANDBYDURATION_250  = 3
	STANDBYDURATION_500  = 4
	STANDBYDURATION_1000 = 5
	STANDBYDURATION_10   = 6
	STANDBYDURATION_20   = 7
)

type StandbyDurationSetting byte

const (
	FILTER_NO = 0
	FILTER_2  = 1
	FILTER_4  = 2
	FILTER_8  = 3
	FILTER_16 = 4
)

type FilterSetting byte

type BME280Config struct {
	Oversample_humidity    Oversample //Oversample //Common for all?
	Oversample_pressure    Oversample //Common for all?
	Oversample_temperature Oversample //Common for all?
	Mode                   DeviceMode
	Standby                StandbyDurationSetting
	Filter                 FilterSetting
	//Forced mode?
}

//In same order as on I2c memory  24bytes
type CalibrationRegs1 struct {
	T1 uint16
	T2 int16
	T3 int16

	P1 uint16
	P2 int16
	P3 int16
	P4 int16
	P5 int16
	P6 int16
	P7 int16
	P8 int16
	P9 int16
}

//9 bytes
type CalibrationRegs2 struct {
	H1 uint8 //Extra points do not matter?
	H2 int16
	H3 uint8
	H4 int16 //Mixed up
	H5 int16 //Mixed up
	H6 int8  //Mixed up
}

func selectI2CSlave(f *os.File, address byte) error {
	//i2c_SLAVE := 0x0703
	_, _, errorcode := syscall.Syscall6(syscall.SYS_IOCTL, f.Fd(), 0x0703, uintptr(address), 0, 0, 0)
	if errorcode != 0 {
		return fmt.Errorf("Select I2C slave errcode %v", errorcode)
	}
	return nil
}

func (p *BME280) WriteReg(address byte, value byte) error {
	err := selectI2CSlave(p.I2CHandle, p.Address)
	if err != nil {
		return err
	}
	_, err = p.I2CHandle.Write([]byte{address, value})
	return err
}

func (p *BME280) ReadRegs(address byte, count byte) ([]byte, error) {
	selectErr := selectI2CSlave(p.I2CHandle, p.Address)
	result := make([]byte, count)
	if selectErr != nil {
		return result, selectErr
	}

	_, err := p.I2CHandle.Write([]byte{address})
	if err != nil {
		return result, err
	}
	_, err = p.I2CHandle.Read(result)
	return result, err
}

/*
Gets open i2c device file (this allows to share same open device file with other i2c devices)
Check ID
read calibration
*/
func (p *BME280) Initialize(f *os.File, i2cAddressBit bool) error {
	if i2cAddressBit {
		p.Address = 0x77
	} else {
		p.Address = 0x76
	}

	p.I2CHandle = f
	arr, err := p.ReadRegs(REGISTER_CALIB00, 24)
	if err != nil {
		return err
	}
	err = binary.Read(bytes.NewReader(arr), binary.LittleEndian, &p.Calib1)
	if err != nil {
		return err
	}
	arr, err = p.ReadRegs(REGISTER_CALIB26, 8)
	if err != nil {
		return err
	}
	//fmt.Printf("Calib 26 arr = %#v\n", arr)
	p.Calib2.H2 = int16(arr[0]) | int16(arr[1])<<8
	p.Calib2.H3 = arr[2]
	p.Calib2.H4 = int16(arr[3])<<4 | int16(arr[4]&0x0F)
	p.Calib2.H5 = int16(arr[5])<<4 | int16(arr[4]&0xF0)>>4
	p.Calib2.H6 = int8(arr[6])
	arr, err = p.ReadRegs(0xA1, 1)
	if err != nil {
		return err
	}
	p.Calib2.H1 = arr[0]
	return nil
}

func (p *BME280) Configure(config BME280Config) error {
	err := p.WriteReg(REGISTER_CTRL_HUM, byte(config.Oversample_humidity))
	if err != nil {
		return err
	}
	err = p.WriteReg(REGISTER_CTRL_MEAS, byte((byte(config.Oversample_temperature)<<5)|(byte(config.Oversample_pressure)<<2)|3)|byte(config.Mode)) //TODO FORCED? or normal?
	if err != nil {
		return err
	}
	//Minimal inactivity, no filter =0. More inactivity, less self heating
	return p.WriteReg(REGISTER_CONFIG, ((byte(config.Standby)&0x7)<<4)|((byte(config.Filter)&0x7)<<1))
}

/*
Reads all results
*/
func (p *BME280) BME280Read() (HumTempPressureMeas, error) {
	var result HumTempPressureMeas

	raw, err := p.ReadRegs(REGISTER_DATA, 8)
	if err != nil {
		return result, err
	}
	traw := uint32(raw[3])<<12 | uint32(raw[4])<<4 | uint32(raw[5])>>4
	praw := uint32(raw[0])<<12 | uint32(raw[1])<<4 | uint32(raw[2])>>4
	hraw := uint32(raw[6])<<8 | uint32(raw[7])

	var v1, v2 float64
	var tfine int32

	v1 = (float64(traw)/16384.0 - float64(p.Calib1.T1)/1024.0) * float64(p.Calib1.T2)
	v2 = (float64(traw)/131072.0 - float64(p.Calib1.T1)/8192.0) * (float64(traw)/131072.0 - float64(p.Calib1.T1)/8192.0) * float64(p.Calib1.T3)
	tfine = int32(v1 + v2)
	result.Temperature = (v1 + v2) / 5120.0

	//----------------
	v1 = float64(tfine)/2.0 - 64000.0
	v2 = v1 * v1 * (float64(p.Calib1.P6) / 32768.0)
	v2 = v2 + v1*(float64(p.Calib1.P5)*2.0)
	v2 = v2/4.0 + (float64(p.Calib1.P4) * 65536.0)
	v1 = (float64(p.Calib1.P3)*v1*v1/524288.0 + float64(p.Calib1.P2)*v1) / 524288.0
	v1 = (1.0 + v1/32768.0) * float64(p.Calib1.P1)

	if v1 == 0 {
		result.Pressure = 0
	} else {
		result.Pressure = 1048576.0 - float64(praw)
		result.Pressure = (result.Pressure - v2/4096.0) * 6250.0 / v1
		v1 = float64(p.Calib1.P9) * result.Pressure * result.Pressure / 2147483648.0
		v2 = result.Pressure * float64(p.Calib1.P8) / 32768.0
		result.Pressure = result.Pressure + (v1+v2+float64(p.Calib1.P7))/16.0
	}
	//---------------------------
	result.Rh = float64(tfine) - 76800.0
	result.Rh = (float64(hraw) - float64(p.Calib2.H4)*64.0 + float64(p.Calib2.H5)/16384.0*result.Rh) * float64(p.Calib2.H2) / 65536.0 * (1.0 + float64(p.Calib2.H6)/67108864.0*result.Rh*(1.0+float64(p.Calib2.H3)/67108864.0*result.Rh))
	result.Rh = result.Rh * (1.0 - float64(p.Calib2.H1)*result.Rh/524288.0)
	return result, nil
}
