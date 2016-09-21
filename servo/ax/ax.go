package ax

import (
	"github.com/adammck/dynamixel/iface"
	"github.com/adammck/dynamixel/protocol/v1"
	reg "github.com/adammck/dynamixel/registers"
	"github.com/adammck/dynamixel/servo"
)

// New returns a new AX-series servo with the given ID.
// See: http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm
func New(n iface.Networker, ID int) (*servo.Servo, error) {
	return servo.New(v1.New(n), Registers, ID), nil
}

var Registers reg.Map

func init() {
	x := 0

	Registers = reg.Map{

		// EEPROM: Persisted
		reg.ModelNumber:             {0x00, 2, reg.RO, x, x},
		reg.FirmwareVersion:         {0x02, 1, reg.RO, x, x},
		reg.ServoID:                 {0x03, 1, reg.RW, 0, 252}, // renamed from ID for clarity
		reg.BaudRate:                {0x04, 1, reg.RW, 0, 254}, // bps = 2000000/(value+1)
		reg.ReturnDelayTime:         {0x05, 1, reg.RW, 0, 254}, // usec = value*2
		reg.CwAngleLimit:            {0x06, 2, reg.RW, 0, 1023},
		reg.CcwAngleLimit:           {0x08, 2, reg.RW, 0, 1023},
		reg.HighestLimitTemperature: {0x0b, 1, reg.RW, 0, 70},   // docs says not to set
		reg.LowestLimitVoltage:      {0x0c, 1, reg.RW, 50, 250}, // volt = value*0.1
		reg.HighestLimitVoltage:     {0x0d, 1, reg.RW, 50, 250}, // volt = value*0.1
		reg.MaxTorque:               {0x0e, 2, reg.RW, 0, 1023}, // from zero to max torque
		reg.StatusReturnLevel:       {0x10, 1, reg.RW, 0, 2},    // enum; see docs
		reg.AlarmLed:                {0x11, 1, reg.RW, 0, 256},  // enum; see docs
		reg.AlarmShutdown:           {0x12, 1, reg.RW, 0, 256},  // enum; see docs

		// RAM: Reset to default when power-cycled
		reg.TorqueEnable:          {0x18, 1, reg.RW, 0, 1},    // bool
		reg.Led:                   {0x19, 1, reg.RW, 0, 1},    // bool
		reg.CwComplianceMargin:    {0x1a, 1, reg.RW, 0, 255},  // def=1
		reg.CcwComplianceMargin:   {0x1b, 1, reg.RW, 0, 255},  // def=1
		reg.CwComplianceSlope:     {0x1c, 1, reg.RW, 0, 254},  // stepped (see docs), def=32
		reg.CcwComplianceSlope:    {0x1d, 1, reg.RW, 0, 254},  // stepped (see docs), def=32
		reg.GoalPosition:          {0x1e, 2, reg.RW, 0, 1023}, // deg = value*0.29; 512 (150 deg) is center
		reg.MovingSpeed:           {0x20, 2, reg.RW, 0, 1023}, // joint mode: rpm = ~value*0.111, but 0 = max rpm. wheel mode: see docs
		reg.TorqueLimit:           {0x22, 2, reg.RW, 0, 1023}, // zero to max torque
		reg.PresentPosition:       {0x24, 2, reg.RO, x, x},    // like goalPosition
		reg.PresentSpeed:          {0x26, 2, reg.RO, x, x},
		reg.PresentLoad:           {0x28, 2, reg.RO, x, x},
		reg.PresentVoltage:        {0x2a, 1, reg.RO, x, x},
		reg.PresentTemperature:    {0x2b, 1, reg.RO, x, x},
		reg.RegisteredInstruction: {0x2c, 1, reg.RO, x, x},
		reg.Moving:                {0x2e, 1, reg.RO, x, x},
		reg.Lock:                  {0x2f, 1, reg.RW, 0, 1}, // bool
		reg.Punch:                 {0x30, 2, reg.RW, 32, 1023},
	}
}
