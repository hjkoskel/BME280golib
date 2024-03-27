//go:build tinygo

package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/hjkoskel/BME280golib"
)

/*
Microcontroller version.

Tested on wio terminal.

Much simpler and hard coded constant values. But at least same code works on microcontroller
*/

func GetDeviceAndParameters() (BME280golib.I2CDeviceLayer, SoftwareParameters, error) {
	i2cConnect := machine.I2C1
	errI2CConfigure := i2cConnect.Configure(machine.I2CConfig{
		SCL: machine.SCL1_PIN,
		SDA: machine.SDA1_PIN,
	})

	if errI2CConfigure != nil {
		return nil, SoftwareParameters{}, errI2CConfigure
	}

	a := BME280golib.CreateI2CTiny(i2cConnect, BME280golib.BME280DEVICEBIT0) //TODO change to BME280golib.BME280DEVICEBIT1 if needed
	ap := &a
	return ap, SoftwareParameters{SensorConf: BME280golib.BME280Config{
		Oversample_humidity:    BME280golib.OVRSAMPLE_1,
		Oversample_pressure:    BME280golib.OVRSAMPLE_1,
		Oversample_temperature: BME280golib.OVRSAMPLE_1,
		Mode:                   BME280golib.MODE_NORMAL,
		Standby:                BME280golib.STANDBYDURATION_500,
		Filter:                 BME280golib.FILTER_NO},
		PollInterval: time.Second,
	}, nil

}

func HandleTermintingError(err error) {
	for err != nil {
		fmt.Printf("FAIL %s\n", err)
		time.Sleep(time.Second * 3)
	}
}
