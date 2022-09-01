/*
Fake sensor with fake configuration

This tries to

*/

package BME280golib

import "fmt"

type Fake280 struct {
	open       bool //Fail if closed
	fakingNow  HumTempPressureMeas
	config     BME280Config
	fakeDataIn chan HumTempPressureMeas
}

func InitializeFake(fakeInput chan HumTempPressureMeas) Fake280 {
	return Fake280{
		open:       true,
		fakeDataIn: fakeInput,
	}
}

func (p *Fake280) Close() error {
	p.open = false
	return nil
}

func (p *Fake280) SoftReset() error {
	return nil
}

func (p *Fake280) Configure(config BME280Config) error {
	if !p.open {
		return fmt.Errorf("configuration failed: closed")
	}
	p.config = config
	return nil
}

//ReadRaw is not really used expect some debugging, testing or research purposes. Use Read
func (p *Fake280) Read() (HumTempPressureMeas, error) {
	if !p.open {
		return HumTempPressureMeas{}, fmt.Errorf("Read failed: closed")
	}

	for 0 < len(p.fakeDataIn) {
		p.fakingNow = <-p.fakeDataIn
	}
	return p.fakingNow, nil

}
