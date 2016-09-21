package xl

import (
	"github.com/adammck/dynamixel/iface"
	"github.com/adammck/dynamixel/protocol/v2"
	reg "github.com/adammck/dynamixel/registers"
	"github.com/adammck/dynamixel/servo"
)

// New returns a new XL-320 servo with the given ID.
// See: http://support.robotis.com/en/product/dynamixel/xl-series/xl-320.htm
func New(n iface.Networker, ID int) (*servo.Servo, error) {
	return servo.New(v2.New(n), Registers, ID), nil
}

func NewWithReturnLevel(n iface.Protocol, ID int, returnLevel int) (*servo.Servo, error) {
	return servo.NewWithReturnLevel(n, Registers, ID, returnLevel), nil
}

var Registers reg.Map

func init() {
	x := 0

	Registers = reg.Map{

		// EEPROM: Persisted
		reg.ModelNumber:             {0x00, 2, reg.RO, x, x},
		reg.FirmwareVersion:         {0x02, 1, reg.RO, x, x},
		reg.ServoID:                 {0x03, 1, reg.RW, 0, 252}, // renamed from ID for clarity
		reg.BaudRate:                {0x04, 1, reg.RW, 0, 3},   // 0=9600, 1=57600, 2=115200, 3=1Mbps
		reg.ReturnDelayTime:         {0x05, 1, reg.RW, 0, 254}, // usec = value*2
		reg.CwAngleLimit:            {0x06, 2, reg.RW, 0, 1023},
		reg.CcwAngleLimit:           {0x08, 2, reg.RW, 0, 1023},
		reg.ControlMode:             {0x0b, 1, reg.RW, 1, 2},    // 1=wheel mode, 2=joint mode
		reg.HighestLimitTemperature: {0x0c, 1, reg.RW, 0, 150},  // docs says not to set
		reg.LowestLimitVoltage:      {0x0d, 1, reg.RW, 50, 250}, // volt = value*0.1
		reg.HighestLimitVoltage:     {0x0e, 1, reg.RW, 50, 250}, // volt = value*0.1
		reg.MaxTorque:               {0x0f, 2, reg.RW, 0, 1023}, // from zero to max torque
		reg.StatusReturnLevel:       {0x11, 1, reg.RW, 0, 2},    // enum; see docs
		reg.AlarmShutdown:           {0x12, 1, reg.RW, 0, 256},  // enum; see docs

		// RAM: Reset to default when power-cycled
		reg.TorqueEnable:          {0x18, 1, reg.RW, 0, 1},
		reg.Led:                   {0x19, 1, reg.RW, 0, 7},
		reg.DGain:                 {0x1b, 1, reg.RW, 0, 254},
		reg.IGain:                 {0x1c, 1, reg.RW, 0, 254},
		reg.PGain:                 {0x1d, 1, reg.RW, 0, 1023},
		reg.GoalPosition:          {0x1e, 2, reg.RW, 0, 1023}, // deg = value*0.29; 512 (150 deg) is center
		reg.GoalVelocity:          {0x20, 2, reg.RW, 0, 2047}, // joint mode: rpm = ~value*0.111, but 0 = max rpm. wheel mode: see docs
		reg.GoalTorque:            {0x23, 2, reg.RW, 0, 1023}, // zero to max torque
		reg.PresentPosition:       {0x25, 2, reg.RO, x, x},    // like goalPosition
		reg.PresentSpeed:          {0x27, 2, reg.RO, x, x},
		reg.PresentLoad:           {0x29, 2, reg.RO, x, x},
		reg.PresentVoltage:        {0x2d, 1, reg.RO, x, x},
		reg.PresentTemperature:    {0x2e, 1, reg.RO, x, x},
		reg.RegisteredInstruction: {0x2f, 1, reg.RO, x, x},
		reg.Moving:                {0x31, 1, reg.RO, x, x},
		reg.HardwareErrorStatus:   {0x32, 1, reg.RO, x, x},
		reg.Punch:                 {0x33, 2, reg.RW, 32, 1023},
	}
}
