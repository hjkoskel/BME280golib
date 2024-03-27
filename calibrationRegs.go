package BME280golib

import (
	"fmt"
	"strings"
)

// CalibrationRegs is combined from CalibrationRegs1 and CalibrationRegs2. For simpler operation
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

// ToTable for printout
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

// In same order as on I2c memory  24bytes
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

// 9 bytes
type CalibrationRegs2 struct {
	H1 uint8 //Extra points do not matter?
	H2 int16
	H3 uint8
	H4 int16 //Mixed up
	H5 int16 //Mixed up
	H6 int8  //Mixed up
}
