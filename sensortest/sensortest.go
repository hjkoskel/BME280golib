/*
Test BME280 sensor.
Simple program for testing functionality and options
*/

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hjkoskel/BME280golib"
)

type SoftwareParameters struct {
	SensorConf   BME280golib.BME280Config
	PollInterval time.Duration
}

type MeasurementResult struct {
	SamplingTime time.Time
	Meas         BME280golib.HumTempPressureMeas
}

func (p *MeasurementResult) ToCsv() string {
	return fmt.Sprintf("%v\t%.1f\t%.3f\t%.3f", p.SamplingTime.UnixMilli(), p.Meas.Pressure, p.Meas.Temperature, p.Meas.Rh)
}

//RunBME is example how to collect results with fixed cycle time

func RunBME(sensor BME280golib.BME280Device, cycleDuration time.Duration, chOut chan<- MeasurementResult) error {
	for {
		tCycleStart := time.Now()
		measResult, measErr := sensor.Read()
		if measErr != nil {
			return measErr
		}
		chOut <- MeasurementResult{Meas: measResult, SamplingTime: time.Now()}

		dur := time.Since(tCycleStart)
		if dur < cycleDuration { //Took less time than required
			time.Sleep(cycleDuration - dur)
		}
	}
}

func main() {
	i2cSensor, pars, parsErr := GetDeviceAndParameters()
	HandleTermintingError(parsErr)

	bmeDevice, errInit := BME280golib.CreateBME280I2C(i2cSensor)
	if errInit != nil {
		HandleTermintingError(fmt.Errorf("Error initializing BME280 %v\n", errInit.Error()))
	}

	resetErr := bmeDevice.SoftReset()
	if resetErr != nil {
		fmt.Printf("Failed on reset %v", resetErr.Error())
		os.Exit(-1)
	}

	confOperErr := bmeDevice.Configure(pars.SensorConf)
	HandleTermintingError(confOperErr)

	if pars.SensorConf.Mode == BME280golib.MODE_SLEEP {
		fmt.Printf("Sleeping....\n")
		time.Sleep(time.Second * 3)
		return
	}

	measDataCh := make(chan MeasurementResult, 100)
	go func() {
		for data := range measDataCh {
			fmt.Printf("%s\n", data.ToCsv())
		}
	}()

	runtimeErr := RunBME(&bmeDevice, pars.PollInterval, measDataCh)
	HandleTermintingError(runtimeErr)
}
