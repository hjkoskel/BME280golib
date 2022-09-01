/*
Test BME280 sensor.
Simple program for testing functionality and options
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
	//fmt.Printf("Ovesample is %v in int %v\n", result, int(result))
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

type MeasurementResult struct {
	SamplingTime time.Time
	Meas         BME280golib.HumTempPressureMeas
}

func (p *MeasurementResult) ToCsv() string {
	return fmt.Sprintf("%v\t%.1f\t%.3f\t%.3f", p.SamplingTime.UnixMilli(), p.Meas.Pressure, p.Meas.Temperature, p.Meas.Rh)
}

//RunBME is example how to collect results with fixed cycle time

func RunBME(sensor *BME280golib.BME280, maxN int, cycleDuration time.Duration, chOut chan<- MeasurementResult) error {
	count := 0
	for count < maxN || maxN == 0 {
		tCycleStart := time.Now()
		measResult, measErr := sensor.Read()
		if measErr != nil {
			return measErr
		}
		chOut <- MeasurementResult{
			Meas:         measResult,
			SamplingTime: time.Now(),
		}

		dur := time.Since(tCycleStart)
		if dur < cycleDuration { //Took less time than required
			time.Sleep(cycleDuration - dur)
		}
		count++
	}
	close(chOut)
	return nil
}

func RunResultProcess(chIn <-chan MeasurementResult) error {
	for data := range chIn {
		fmt.Printf("%s\n", data.ToCsv())
	}
	return nil
}

func main() {
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
	i2cDeviceFileHandle, errI2CHardware := os.OpenFile(swpars.DeviceFileName, os.O_RDWR, 0600)
	if errI2CHardware != nil {
		fmt.Printf("I2C error on device file error=%v", errI2CHardware.Error())
		os.Exit(-1)
	}
	bmeDevice, errInit := BME280golib.InitBME280I2C(i2cDeviceFileHandle, swpars.I2CBit)
	if errInit != nil {
		fmt.Printf("Error initializing BME280 %v\n", errInit.Error())
		os.Exit(-1)
	}

	if swpars.SoftResetStart {
		resetErr := bmeDevice.SoftReset()
		if resetErr != nil {
			fmt.Printf("Failed on reset %v", resetErr.Error())
			os.Exit(-1)
		}
	}

	//CONFIGURE
	confOperErr := bmeDevice.Configure(conf)
	if confOperErr != nil {
		fmt.Printf("Device configuration failed %v", confOperErr.Error())
	}

	if conf.Mode == BME280golib.MODE_SLEEP {
		fmt.Printf("Sleeping....\n")
		return
	}
	if conf.Mode == BME280golib.MODE_FORCED {
		fmt.Printf("Forced mode, one sample only (needs re-set mode)\n")
		swpars.SampleCount = 1
	}

	measDataCh := make(chan MeasurementResult, 1000)
	go func() {
		processErr := RunResultProcess(measDataCh)
		if processErr != nil {
			fmt.Printf("Result processing failed %v\n", processErr.Error())
			os.Exit(-1)
		}
	}()

	runtimeErr := RunBME(&bmeDevice, swpars.SampleCount, cycleDuration, measDataCh)
	if runtimeErr != nil {
		fmt.Printf("\n\nRuntime error %v\n", runtimeErr.Error())
		os.Exit(-1)
	}
}
