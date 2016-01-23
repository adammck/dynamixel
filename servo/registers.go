package servo

import (
	"github.com/adammck/dynamixel"
)

const (

	// Register names are used to refer to a specific value in a servo's EEPROM or
	// RAM. See the registers variable for the addresses and lengths.
	modelNumber dynamixel.RegName = iota
	firmwareVersion
	servoID
	baudRate
	returnDelayTime
	cwAngleLimit
	ccwAngleLimit
	highestLimitTemperature
	lowestLimitVoltage
	highestLimitVoltage
	maxTorque
	statusReturnLevel
	alarmLed
	alarmShutdown
	torqueEnable
	led
	cwComplianceMargin
	ccwComplianceMargin
	cwComplianceSlope
	ccwComplianceSlope
	goalPosition
	movingSpeed
	torqueLimit
	presentPosition
	presentSpeed
	presentLoad
	presentVoltage
	presentTemperature
	registered
	moving
	lock
	punch
)

// See: http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm
var registers map[dynamixel.RegName]*dynamixel.Register

func init() {
	x := 0
	ro := dynamixel.RO
	rw := dynamixel.RW

	registers = map[dynamixel.RegName]*dynamixel.Register{

		// EEPROM: Persisted
		modelNumber:             {0x00, 2, ro, true, x, x},
		firmwareVersion:         {0x02, 1, ro, true, x, x},
		servoID:                 {0x03, 1, rw, true, 0, 252}, // renamed from ID for clarity
		baudRate:                {0x04, 1, rw, true, 0, 254}, // bps = 2000000/(value+1)
		returnDelayTime:         {0x05, 1, rw, true, 0, 254}, // usec = value*2
		cwAngleLimit:            {0x06, 2, rw, true, 0, 1023},
		ccwAngleLimit:           {0x08, 2, rw, true, 0, 1023},
		highestLimitTemperature: {0x0b, 1, rw, true, 0, 70},   // docs says not to set
		lowestLimitVoltage:      {0x0c, 1, rw, true, 50, 250}, // volt = value*0.1
		highestLimitVoltage:     {0x0d, 1, rw, true, 50, 250}, // volt = value*0.1
		maxTorque:               {0x0e, 2, rw, true, 0, 1023}, // from zero to max torque
		statusReturnLevel:       {0x10, 1, rw, true, 0, 2},    // enum; see docs
		alarmLed:                {0x11, 1, rw, true, 0, 256},  // enum; see docs
		alarmShutdown:           {0x12, 1, rw, true, 0, 256},  // enum; see docs

		// RAM: Reset to default when power-cycled
		torqueEnable:        {0x18, 1, rw, true, 0, 1},    // bool
		led:                 {0x19, 1, rw, true, 0, 1},    // bool
		cwComplianceMargin:  {0x1a, 1, rw, true, 0, 255},  // def=1
		ccwComplianceMargin: {0x1b, 1, rw, true, 0, 255},  // def=1
		cwComplianceSlope:   {0x1c, 1, rw, true, 0, 254},  // stepped (see docs), def=32
		ccwComplianceSlope:  {0x1d, 1, rw, true, 0, 254},  // stepped (see docs), def=32
		goalPosition:        {0x1e, 2, rw, true, 0, 1023}, // deg = value*0.29; 512 (150 deg) is center
		movingSpeed:         {0x20, 2, rw, true, 0, 1023}, // joint mode: rpm = ~value*0.111, but 0 = max rpm. wheel mode: see docs
		torqueLimit:         {0x22, 2, rw, true, 0, 1023}, // zero to max torque
		presentPosition:     {0x24, 2, ro, false, x, x},   // like goalPosition
		presentSpeed:        {0x26, 2, ro, true, x, x},
		presentLoad:         {0x28, 2, ro, true, x, x},
		presentVoltage:      {0x2a, 1, ro, false, x, x},
		presentTemperature:  {0x2b, 1, ro, true, x, x},
		registered:          {0x2c, 1, ro, true, x, x},
		moving:              {0x2e, 1, ro, true, x, x},
		lock:                {0x2f, 1, rw, true, 0, 1}, // bool
		punch:               {0x30, 2, rw, true, 32, 1023},
	}
}
