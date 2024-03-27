//go:build !tinygo

/*
linux specific things on this test program
Built with normal golang compiler
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hjkoskel/BME280golib"
)

func IntToOversample(i int) (BME280golib.Oversample, error) {
	m := map[int]BME280golib.Oversample{
		0:  BME280golib.OVRSAMPLE_NO,
		1:  BME280golib.OVRSAMPLE_1,
		2:  BME280golib.OVRSAMPLE_2,
		4:  BME280golib.OVRSAMPLE_4,
		8:  BME280golib.OVRSAMPLE_8,
		16: BME280golib.OVRSAMPLE_16,
	}

	result, haz := m[i]
	if !haz {
		return result, fmt.Errorf("Invalid oversample %v, use 0,1,2,4,8 or 16", i)
	}
	return result, nil
}

func IntToFilter(i int) (BME280golib.FilterSetting, error) {
	m := map[int]BME280golib.FilterSetting{
		0:  BME280golib.FILTER_NO,
		2:  BME280golib.FILTER_2,
		4:  BME280golib.FILTER_4,
		8:  BME280golib.FILTER_8,
		16: BME280golib.FILTER_16}
	result, haz := m[i]
	if !haz {
		return result, fmt.Errorf("Invalid filter %v, use 0,2,4,8 or 16", i)
	}
	return result, nil
}

type SwParameters struct {
	RequestedInterval time.Duration
	I2CBit            bool
	DeviceFileName    string
	SoftResetStart    bool
	SampleCount       int
}

func GetParsFromFlags() (BME280golib.BME280Config, SwParameters, error) {
	pHov := flag.Int64("hov", 1, "Oversampling humidity 0=no, 1,2,4,8 or 16")
	pTov := flag.Int64("tov", 1, "Oversampling temperature 0=no, 1,2,4,8 or 16")
	pPov := flag.Int64("pov", 1, "Oversampling pressure 0=no, 1,2,4,8 or 16")
	pMode := flag.Int64("mode", 3, "idle=0, forced=1, normal=3")
	pInterval := flag.Int64("iv", 0, "requested reading interval in milliseconds")
	pStandbyTime := flag.Float64("sb", 1000, "standby duration milliseconds")
	pFilter := flag.Int64("filt", 0, "filter 0,2,4,8 or 16")
	pI2CaddressBit := flag.Bool("i2cbit", false, "select I2C address false=0x76, true=0x77")
	pI2CDeviceFile := flag.String("i2cdev", "/dev/i2c-1", "I2C device file")
	pSoftReset := flag.Bool("softreset", false, "do soft reset at start")
	pN := flag.Int64("n", 100, "Number of samples measured n=0 is forever")
	flag.Parse()

	swPars := SwParameters{
		RequestedInterval: time.Duration(*pInterval) * time.Millisecond,
		I2CBit:            *pI2CaddressBit,
		DeviceFileName:    *pI2CDeviceFile,
		SoftResetStart:    *pSoftReset,
		SampleCount:       int(*pN),
	}

	conf := BME280golib.BME280Config{}

	var errPar error
	conf.Oversample_humidity, errPar = IntToOversample(int(*pHov))
	if errPar != nil {
		return conf, swPars, errPar
	}
	conf.Oversample_temperature, errPar = IntToOversample(int(*pTov))
	if errPar != nil {
		return conf, swPars, errPar
	}
	conf.Oversample_pressure, errPar = IntToOversample(int(*pPov))
	if errPar != nil {
		return conf, swPars, errPar
	}

	sb := time.Duration(*pStandbyTime*1000) * time.Microsecond
	fmt.Printf("standby time given %v\n", sb)
	conf.Standby = BME280golib.GetStandbyDuration(sb)
	conf.Mode = BME280golib.DeviceMode(*pMode)
	conf.Filter, errPar = IntToFilter(int(*pFilter))
	if errPar != nil {
		return conf, swPars, errPar
	}
	return conf, swPars, conf.GotError()
}

func GetDeviceAndParameters() (BME280golib.I2CDeviceLayer, SoftwareParameters, error) {
	conf, swpars, confErr := GetParsFromFlags()
	if confErr != nil {
		fmt.Printf("ERROR %v\n", confErr.Error())
		os.Exit(-1)
	}

	fmt.Printf("\n--Configuration --\n%s\n\nMeasurement duration  typ=%v max=%v\n", conf,
		conf.MeasurementDurationTypical(),
		conf.MeasurementDurationMaximum())

	fmt.Printf("Software parameters %#v\n", swpars)
	cycleDuration := conf.CycleDuration()
	if cycleDuration < swpars.RequestedInterval {
		cycleDuration = swpars.RequestedInterval
	}
	fmt.Printf("Running %v long measurement cycle\n", cycleDuration)

	//INIT device
	i2cAddress := BME280golib.BME280DEVICEBIT0
	if swpars.I2CBit {
		i2cAddress = BME280golib.BME280DEVICEBIT1
	}

	i2cFile, errOpenI2cFile := os.OpenFile(swpars.DeviceFileName, os.O_RDWR, 0600)
	if errOpenI2cFile != nil {
		fmt.Printf("error opening I2C device file %s  err=%s\n", swpars.DeviceFileName, errOpenI2cFile)
		os.Exit(-1)
	}

	a := BME280golib.CreateI2CSys(i2cFile, i2cAddress)
	return &a, SoftwareParameters{
		SensorConf:   conf,
		PollInterval: cycleDuration}, nil
}

func HandleTermintingError(err error) {
	if err != nil {
		fmt.Printf("ERROR %s\n", err.Error())
		os.Exit(-1)
	}
}
