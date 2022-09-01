package BME280golib

import "math"

//RawMeas is what registers have, needs calibration (stored on chip) for creating correct readout HumTempPressureMeas
type RawMeas struct {
	/*
		Pmsb  byte
		Plsb  byte
		Pxlsb byte

		Tmsb  byte
		Tlsb  byte
		Txlsb byte

		Hmsb byte
		Hlsb byte

		traw := uint32(raw[3])<<12 | uint32(raw[4])<<4 | uint32(raw[5])>>4
		praw := uint32(raw[0])<<12 | uint32(raw[1])<<4 | uint32(raw[2])>>4
		hraw := uint32(raw[6])<<8 | uint32(raw[7])
	*/

	Temperature uint32 //24bit
	Pressure    uint32 //24bit
	Humidity    uint16 //16bit
}

type HumTempPressureMeas struct {
	Temperature float64
	Rh          float64
	Pressure    float64
}

//Compensate, with by datasheet. 8.1 Compensation formulas in double precision floating point
func (p *RawMeas) Compensate(calib CalibrationRegs) (HumTempPressureMeas, error) {
	var v1, v2 float64
	var tfine int32

	result := HumTempPressureMeas{}

	v1 = (float64(p.Temperature)/16384.0 - float64(calib.T1)/1024.0) * float64(calib.T2)
	v2 = (float64(p.Temperature)/131072.0 - float64(calib.T1)/8192.0) * (float64(p.Temperature)/131072.0 - float64(calib.T1)/8192.0) * float64(calib.T3)
	tfine = int32(v1 + v2)
	result.Temperature = (v1 + v2) / 5120.0

	//----------------
	v1 = float64(tfine)/2.0 - 64000.0
	v2 = v1 * v1 * (float64(calib.P6) / 32768.0)
	v2 = v2 + v1*(float64(calib.P5)*2.0)
	v2 = v2/4.0 + (float64(calib.P4) * 65536.0)
	v1 = (float64(calib.P3)*v1*v1/524288.0 + float64(calib.P2)*v1) / 524288.0
	v1 = (1.0 + v1/32768.0) * float64(calib.P1)

	if v1 == 0 {
		result.Pressure = 0
	} else {
		result.Pressure = 1048576.0 - float64(p.Pressure)
		result.Pressure = (result.Pressure - v2/4096.0) * 6250.0 / v1
		v1 = float64(calib.P9) * result.Pressure * result.Pressure / 2147483648.0
		v2 = result.Pressure * float64(calib.P8) / 32768.0
		result.Pressure = result.Pressure + (v1+v2+float64(calib.P7))/16.0
	}
	//---------------------------
	result.Rh = float64(tfine) - 76800.0
	result.Rh = (float64(p.Humidity) - float64(calib.H4)*64.0 + float64(calib.H5)/16384.0*result.Rh) * float64(calib.H2) / 65536.0 * (1.0 + float64(calib.H6)/67108864.0*result.Rh*(1.0+float64(calib.H3)/67108864.0*result.Rh))
	result.Rh = result.Rh * (1.0 - float64(calib.H1)*result.Rh/524288.0)

	result.DoInfs()
	return result, nil
}

func (p *HumTempPressureMeas) AbsDiff(a HumTempPressureMeas) HumTempPressureMeas {
	return HumTempPressureMeas{Temperature: math.Abs(p.Temperature - a.Temperature), Rh: math.Abs(p.Rh - a.Rh), Pressure: math.Abs(p.Pressure - a.Pressure)}
}

/*
On register level
Temperature and Rh are 24bit. Pressure is 16bit register?

Pressure range 30kPa to 110kPa, absolute accuracy 0.1kPa
Best resolution with oversampling 0.18Pa = 0.00018 ->  444444 steps   19bits?
Reduced bandwidth noise floor 0.2Pa  -> 0.0002 kPa  40000 steps ->  16bit enough


Temperature -40 to 85  (full accuracy 0 to 65)
best accuracy +-0.5Celsius  -> 250steps  8bits?
Api output resolution 0.01	-> 14bits

Humidity 0 to 100. abs accuracy  +-3,
Resolution 0.0008 -> 12500


*/

//Maximum limits for sanity checking
const (
	PRESSURE_MIN float64 = 30000
	PRESSURE_MAX float64 = 110000

	TEMPERATURE_MIN float64 = -40
	TEMPERATURE_MAX float64 = 85

	HUMIDITY_MIN float64 = 0
	HUMIDITY_MAX float64 = 100
)

//Makes infs to results if out of measurement range. Some people do not agree this.
func (p *HumTempPressureMeas) DoInfs() {
	if p.Pressure < PRESSURE_MIN {
		p.Pressure = math.Inf(-1)
	}
	if PRESSURE_MAX < p.Pressure {
		p.Pressure = math.Inf(1)
	}

	if p.Temperature < TEMPERATURE_MIN {
		p.Temperature = math.Inf(-1)
	}
	if TEMPERATURE_MAX < p.Temperature {
		p.Temperature = math.Inf(1)
	}

	if p.Rh < HUMIDITY_MIN {
		p.Rh = math.Inf(-1)
	}
	if HUMIDITY_MAX < p.Rh {
		p.Rh = math.Inf(1)
	}
}
