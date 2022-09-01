/*
Library for reading BME280 from i2c device file
*/
package BME280golib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

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
}

type BME280 struct {
	i2cHandle *os.File //Set this at startup
	//TODO SPI version? But I do not have SPI breakout board available

	Calib   CalibrationRegs
	address byte
}

//CalibrationRegs is combined from CalibrationRegs1 and CalibrationRegs2. For simpler operation
type CalibrationRegs struct {
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

	H1 uint8 //Extra points do not matter?
	H2 int16
	H3 uint8
	H4 int16 //Mixed up
	H5 int16 //Mixed up
	H6 int8  //Mixed up
}

func CombineCalibrations(c1 CalibrationRegs1, c2 CalibrationRegs2) CalibrationRegs {
	return CalibrationRegs{
		T1: c1.T1,
		T2: c1.T2,
		T3: c1.T3,

		P1: c1.P1,
		P2: c1.P2,
		P3: c1.P3,
		P4: c1.P4,
		P5: c1.P5,
		P6: c1.P6,
		P7: c1.P7,
		P8: c1.P8,
		P9: c1.P9,

		H1: c2.H1,
		H2: c2.H2,
		H3: c2.H3,
		H4: c2.H4,
		H5: c2.H5,
		H6: c2.H6,
	}
}

//ToTable for printout
func (a CalibrationRegs) ToTable() string {
	var sb strings.Builder
	sb.WriteString("Reg\tPressure\tHumidity\tTemperature\n")
	sb.WriteString(fmt.Sprintf("1\t%v\t%v\t%v\n2\t%v\t%v\t%v\n3\t%v\t%v\t%v\n",
		a.P1, a.H1, a.T1,
		a.P2, a.H2, a.T2,
		a.P3, a.H3, a.T3,
	))
	sb.WriteString(fmt.Sprintf("4\t%v\t%v\t\n5\t%v\t%v\t\n6\t%v\t%v\t\n",
		a.P4, a.H4,
		a.P5, a.H5,
		a.P6, a.H6))

	sb.WriteString(fmt.Sprintf("7\t%v\t\t\n8\t%v\t\t\n9\t%v\t\t\n",
		a.P7, a.P8, a.P9))
	return sb.String()
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

/*
Gets open i2c device file (this allows to share same open device file with other i2c devices)
Check ID
read calibration

TODO IMPLEMENT SPI support when such hardware is available/need
*/
func InitBME280I2C(f *os.File, i2cAddressBit bool) (BME280, error) {
	result := BME280{address: 0x76, i2cHandle: f}
	if i2cAddressBit {
		result.address = 0x77
	}
	errInit := result.init()
	if errInit != nil {
		return BME280{}, errInit
	}
	return result, nil
}

//Common initialization for I2C and SPI
func (p *BME280) init() error {
	//Check ID
	idArr, idErrRead := p.readRegs(REGISTER_ID, 1)
	if idErrRead != nil {
		return fmt.Errorf("Id check read error %v\n", idErrRead.Error())
	}

	if idArr[0] != ID_EXPECTED {
		return fmt.Errorf("Invalid BME280 id=0x%02X expected 0x%02X", idArr[0], ID_EXPECTED)
	}

	var calibErr error
	p.Calib, calibErr = p.readCalibration()
	if calibErr != nil {
		return fmt.Errorf("Reading calibration failed %v", calibErr.Error())
	}
	//TODO: calibration sanity check? no zeros or ones only
	return nil
}

//readCalibration reads calibration once, for that reason this is private
func (p *BME280) readCalibration() (CalibrationRegs, error) {
	var calib1 CalibrationRegs1
	var calib2 CalibrationRegs2

	arr, err := p.readRegs(REGISTER_CALIB00, 24)
	if err != nil {
		return CalibrationRegs{}, err
	}
	err = binary.Read(bytes.NewReader(arr), binary.LittleEndian, &calib1)
	if err != nil {
		return CalibrationRegs{}, err
	}
	arr, err = p.readRegs(REGISTER_CALIB26, 8)
	if err != nil {
		return CalibrationRegs{}, err
	}
	//fmt.Printf("Calib 26 arr = %#v\n", arr)
	calib2.H2 = int16(arr[0]) | int16(arr[1])<<8
	calib2.H3 = arr[2]
	calib2.H4 = int16(arr[3])<<4 | int16(arr[4]&0x0F)
	calib2.H5 = int16(arr[5])<<4 | int16(arr[4]&0xF0)>>4
	calib2.H6 = int8(arr[6])
	arr, err = p.readRegs(0xA1, 1)
	if err != nil {
		return CalibrationRegs{}, err
	}
	calib2.H1 = arr[0]
	return CombineCalibrations(calib1, calib2), nil
}

func (p *BME280) Close() error {
	return p.i2cHandle.Close()
}

func (p *BME280) Configure(config BME280Config) error {
	err := p.writeReg(REGISTER_CTRL_HUM, byte(config.Oversample_humidity))
	if err != nil {
		return err
	}

	ctrlByte := byte(config.Oversample_temperature)<<5 | byte(config.Oversample_pressure)<<2 | byte(config.Mode)
	fmt.Printf("ctrlbyte is %v\n", ctrlByte)
	err = p.writeReg(REGISTER_CTRL_MEAS, ctrlByte) //TODO FORCED? or normal?
	if err != nil {
		return err
	}
	//Minimal inactivity, no filter =0. More inactivity, less self heating
	return p.writeReg(REGISTER_CONFIG, ((byte(config.Standby)&0x7)<<4)|((byte(config.Filter)&0x7)<<1))
}

//does soft reset, for glitch etc.... after that re-write configuration
func (p *BME280) SoftReset() error {
	err := p.writeReg(REGISTER_RESET, 0xB6)
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)
	return nil
}

//BME280ReadRaw gives non-compensated readout not really used expect some debugging, testing or research purposes. Use Read
func (p *BME280) ReadRaw() (RawMeas, error) {
	raw, err := p.readRegs(REGISTER_DATA, 8)
	if err != nil {
		return RawMeas{}, err
	}
	return RawMeas{
		Temperature: uint32(raw[3])<<12 | uint32(raw[4])<<4 | uint32(raw[5])>>4,
		Pressure:    uint32(raw[0])<<12 | uint32(raw[1])<<4 | uint32(raw[2])>>4,
		Humidity:    uint16(raw[6])<<8 | uint16(raw[7])}, nil
}

//BME280Read() Reads all results and do internal compensation This is how usually this is used
func (p *BME280) Read() (HumTempPressureMeas, error) {
	raw, rawErr := p.ReadRaw()
	if rawErr != nil {
		return HumTempPressureMeas{}, rawErr
	}
	return raw.Compensate(p.Calib)
}
