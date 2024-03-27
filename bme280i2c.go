package BME280golib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	BME280DEVICEBIT0 uint16 = 0x76
	BME280DEVICEBIT1 uint16 = 0x77
)

type BME280I2C struct {
	dev   I2CDeviceLayer
	calib CalibrationRegs
}

/*
Gets open i2c device file (this allows to share same open device file with other i2c devices)
Check ID
read calibration

TODO IMPLEMENT SPI support when such hardware is available/need
*/
func CreateBME280I2C(layer I2CDeviceLayer) (BME280I2C, error) {
	result := BME280I2C{dev: layer}
	//Check ID
	idArr, idErrRead := result.dev.ReadRegs(REGISTER_ID, 1)
	if idErrRead != nil {
		return result, fmt.Errorf("id check read error %v", idErrRead.Error())
	}

	if idArr[0] != ID_EXPECTED {
		return result, fmt.Errorf("invalid BME280 id=0x%02X expected 0x%02X", idArr[0], ID_EXPECTED)
	}

	var calibErr error
	result.calib, calibErr = result.readCalibration()
	if calibErr != nil {
		return result, fmt.Errorf("reading calibration failed %v", calibErr.Error())
	}
	//TODO: calibration sanity check? no zeros or ones only
	return result, nil
}

func (p *BME280I2C) GetCalibration() (CalibrationRegs, error) {
	calib, errRead := p.readCalibration()
	if errRead != nil {
		return calib, errRead
	}
	p.calib = calib
	return calib, nil

}

// readCalibration reads calibration once, for that reason this is private
func (p *BME280I2C) readCalibration() (CalibrationRegs, error) {
	var calib1 CalibrationRegs1
	var calib2 CalibrationRegs2

	arr, err := p.dev.ReadRegs(REGISTER_CALIB00, 24)
	if err != nil {
		return CalibrationRegs{}, err
	}
	err = binary.Read(bytes.NewReader(arr), binary.LittleEndian, &calib1)
	if err != nil {
		return CalibrationRegs{}, err
	}
	arr, err = p.dev.ReadRegs(REGISTER_CALIB26, 8)
	if err != nil {
		return CalibrationRegs{}, err
	}

	calib2.H2 = int16(arr[0]) | int16(arr[1])<<8
	calib2.H3 = arr[2]
	calib2.H4 = int16(arr[3])<<4 | int16(arr[4]&0x0F)
	calib2.H5 = int16(arr[5])<<4 | int16(arr[4]&0xF0)>>4
	calib2.H6 = int8(arr[6])
	arr, err = p.dev.ReadRegs(0xA1, 1)
	if err != nil {
		return CalibrationRegs{}, err
	}
	calib2.H1 = arr[0]
	return CombineCalibrations(calib1, calib2), nil
}

func (p *BME280I2C) Close() error {
	return p.dev.Close()
}

func (p *BME280I2C) Configure(config BME280Config) error {
	err := p.dev.WriteReg(REGISTER_CTRL_HUM, byte(config.Oversample_humidity))
	if err != nil {
		return err
	}

	ctrlByte := byte(config.Oversample_temperature)<<5 | byte(config.Oversample_pressure)<<2 | byte(config.Mode)
	err = p.dev.WriteReg(REGISTER_CTRL_MEAS, ctrlByte)
	if err != nil {
		return err
	}
	//Minimal inactivity, no filter =0. More inactivity, less self heating
	return p.dev.WriteReg(REGISTER_CONFIG, ((byte(config.Standby)&0x7)<<4)|((byte(config.Filter)&0x7)<<1))
}

// does soft reset, for glitch etc.... after that re-write configuration
func (p *BME280I2C) SoftReset() error {
	err := p.dev.WriteReg(REGISTER_RESET, 0xB6)
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)
	return nil
}

// BME280ReadRaw gives non-compensated readout not really used expect some debugging, testing or research purposes. Use Read
func (p *BME280I2C) ReadRaw() (RawMeas, error) {
	raw, err := p.dev.ReadRegs(REGISTER_DATA, 8)
	if err != nil {
		return RawMeas{}, err
	}
	return RawMeas{
		Temperature: uint32(raw[3])<<12 | uint32(raw[4])<<4 | uint32(raw[5])>>4,
		Pressure:    uint32(raw[0])<<12 | uint32(raw[1])<<4 | uint32(raw[2])>>4,
		Humidity:    uint16(raw[6])<<8 | uint16(raw[7])}, nil
}

// BME280Read() Reads all results and do internal compensation This is how usually this is used
func (p *BME280I2C) Read() (HumTempPressureMeas, error) {
	raw, rawErr := p.ReadRaw()
	if rawErr != nil {
		return HumTempPressureMeas{}, rawErr
	}
	return raw.Compensate(p.calib)
}
