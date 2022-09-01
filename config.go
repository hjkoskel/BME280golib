package BME280golib

import (
	"fmt"
	"time"
)

/*
Oversample setting
*/
type Oversample byte //3bits
const (
	OVRSAMPLE_NO Oversample = 0
	OVRSAMPLE_1  Oversample = 1
	OVRSAMPLE_2  Oversample = 2
	OVRSAMPLE_4  Oversample = 3
	OVRSAMPLE_8  Oversample = 4
	OVRSAMPLE_16 Oversample = 5
)

func (a Oversample) HowManyTimes() int {
	m := map[Oversample]int{
		OVRSAMPLE_NO: 0,
		OVRSAMPLE_1:  1,
		OVRSAMPLE_2:  2,
		OVRSAMPLE_4:  4,
		OVRSAMPLE_8:  8,
		OVRSAMPLE_16: 16}
	result, haz := m[a]
	if !haz {
		return -1
	}
	return result
}

func (a Oversample) String() string {
	if 7 < a {
		return "invalid over 3bit"
	}

	if a == OVRSAMPLE_NO {
		return "skipped"
	}
	if OVRSAMPLE_16 < a {
		return "16x" //Others
	}
	return fmt.Sprintf("%vx", a.HowManyTimes())
}

/*
Device model setting
*/
type DeviceMode byte

const (
	MODE_SLEEP  DeviceMode = 0
	MODE_FORCED DeviceMode = 1
	MODE_NORMAL DeviceMode = 3
)

func (a DeviceMode) String() string {
	switch a {
	case MODE_SLEEP:
		return "sleep"
	case MODE_FORCED:
		return "forced"
	case MODE_NORMAL:
		return "normal"
	default:
		return "INVALID"
	}
}

/*
Standby duration
*/
type StandbyDurationSetting byte

const (
	STANDBYDURATION_0_5  StandbyDurationSetting = 0
	STANDBYDURATION_62_5 StandbyDurationSetting = 1
	STANDBYDURATION_125  StandbyDurationSetting = 2
	STANDBYDURATION_250  StandbyDurationSetting = 3
	STANDBYDURATION_500  StandbyDurationSetting = 4
	STANDBYDURATION_1000 StandbyDurationSetting = 5
	STANDBYDURATION_10   StandbyDurationSetting = 6
	STANDBYDURATION_20   StandbyDurationSetting = 7
)

var standbyDurOptionMicroseconds = []int{500, 62500, 125000, 250000, 500000, 1000000, 10000, 20000}

//GetStandbyDuration picks neareste setting
func GetStandbyDuration(dur time.Duration) StandbyDurationSetting { //Get nearest setting

	result := int(0)
	minerr := int(9999999)
	d := int(dur.Microseconds())

	if 1000000 <= d {
		return STANDBYDURATION_1000
	}

	for i, micros := range standbyDurOptionMicroseconds {
		delta := d - micros
		if delta < 0 {
			delta = -delta
		}
		if delta < minerr {
			minerr = delta
			result = i
		}
	}
	return StandbyDurationSetting(result)

	/*	for i, micros := range standbyDurOptionMicroseconds {
			d := int(dur.Microseconds())
			if d <= micros {
				if i == 0 {
					return StandbyDurationSetting(0)
				}
				prev := standbyDurOptionMicroseconds[i-1]
				if (d - prev) < (micros - d) {
					return StandbyDurationSetting(i - 1)
				}
				return StandbyDurationSetting(i)
			}
		}
		return StandbyDurationSetting(len(standbyDurOptionMicroseconds) - 1)
	*/
}

//Duration how much in between reads. samplerate=1/(standbyduration+readTimes) in normal mode
func (a StandbyDurationSetting) Duration() time.Duration {
	if STANDBYDURATION_20 < a {
		return time.Duration(0) //INVALID
	}
	return time.Duration(standbyDurOptionMicroseconds[int(a)]) * time.Microsecond
}

func (a StandbyDurationSetting) String() string {
	dur := a.Duration()
	if dur == time.Duration(0) {
		return "INVALID"
	}
	return dur.String()
}

/*
Filter setting
*/
type FilterSetting byte

const (
	FILTER_NO FilterSetting = 0
	FILTER_2  FilterSetting = 1
	FILTER_4  FilterSetting = 2
	FILTER_8  FilterSetting = 3
	FILTER_16 FilterSetting = 4
)

func (a FilterSetting) String() string {
	switch a {
	case FILTER_NO:
		return "no"
	case FILTER_2:
		return "2"
	case FILTER_4:
		return "4"
	case FILTER_8:
		return "8"
	case FILTER_16:
		return "16"
	}
	return "invalid"
}

/*
TODO include step response table (figure 7 at manual) here
*/

type BME280Config struct {
	Oversample_humidity    Oversample //Oversample //Common for all?
	Oversample_pressure    Oversample //Common for all?
	Oversample_temperature Oversample //Common for all?
	Mode                   DeviceMode
	Standby                StandbyDurationSetting
	Filter                 FilterSetting
	//Forced mode?
}

func (a BME280Config) String() string {
	return fmt.Sprintf("mode:%s, filter:%s, standby:%v, oversampling:(hum %s,pre %v,temp %v)",
		a.Mode, a.Filter, a.Standby, a.Oversample_humidity, a.Oversample_pressure, a.Oversample_temperature)
}

//GotError check is there bad configuration
func (p *BME280Config) GotError() error {
	if 7 < p.Oversample_humidity {
		return fmt.Errorf("humidity oversample over 3bits")
	}
	if 7 < p.Oversample_pressure {
		return fmt.Errorf("pressure oversample over 3bits")
	}
	if 7 < p.Oversample_pressure {
		return fmt.Errorf("pressure oversample over 3bits")
	}

	if p.Mode != MODE_NORMAL && p.Mode != MODE_FORCED && p.Mode != MODE_SLEEP {
		return fmt.Errorf("invalid mode")
	}
	if 7 < p.Standby {
		return fmt.Errorf("standby over 3bits")
	}

	if 7 < p.Filter {
		return fmt.Errorf("filter over 3bits")
	}

	return nil
}

/*
TODO calculate estimated sampling rate from config?
*/

func (p *BME280Config) MeasurementDurationTypical() time.Duration {
	result := time.Millisecond
	if p.Oversample_temperature != OVRSAMPLE_NO {
		result += time.Millisecond * time.Duration(2*p.Oversample_temperature.HowManyTimes())
	}
	if p.Oversample_pressure != OVRSAMPLE_NO {
		result += time.Microsecond*500 + time.Millisecond*time.Duration(2*p.Oversample_pressure.HowManyTimes())
	}
	if p.Oversample_humidity != OVRSAMPLE_NO {
		result += time.Microsecond*500 + time.Millisecond*time.Duration(2*p.Oversample_humidity.HowManyTimes())
	}

	if p.Mode == MODE_NORMAL {
		result += p.Standby.Duration()
	}

	return result
}

func (p *BME280Config) MeasurementDurationMaximum() time.Duration {
	result := 1250 * time.Microsecond
	if p.Oversample_temperature != OVRSAMPLE_NO {
		result += time.Microsecond * time.Duration(2300*p.Oversample_temperature.HowManyTimes())
	}
	if p.Oversample_pressure != OVRSAMPLE_NO {
		result += time.Microsecond*575 + time.Microsecond*time.Duration(2300*p.Oversample_pressure.HowManyTimes())
	}
	if p.Oversample_humidity != OVRSAMPLE_NO {
		result += time.Microsecond*575 + time.Microsecond*time.Duration(2300*p.Oversample_humidity.HowManyTimes())
	}
	return result
}

func (p *BME280Config) CycleDuration() time.Duration {
	result := p.MeasurementDurationMaximum()
	if p.Mode == MODE_NORMAL { //Controlled by
		result += p.Standby.Duration()
	}
	return result
}
