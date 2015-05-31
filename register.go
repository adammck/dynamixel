package dynamixel

type RegName int
type Access int

const (

	// Register names are used to refer to a specific value in a servo's EEPROM or
	// RAM. See the registers variable for the addresses and lengths.
	modelNumber RegName = iota
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

	// Access Levels specify whether a register is hard-coded into the servo (e.g.
	// the model number), or is a value which can be changed (e.g. the identity).
	ro Access = iota
	rw
)

type Register struct {
	address   byte
	length    int
	access    Access
	cacheable bool
}

var registers map[RegName]*Register

func init() {
	registers = map[RegName]*Register{
		modelNumber:             {0x00, 2, ro, true},
		firmwareVersion:         {0x02, 1, ro, true},
		servoID:                 {0x03, 1, rw, true}, // renamed from ID for clarity
		baudRate:                {0x04, 1, rw, true},
		returnDelayTime:         {0x05, 1, rw, true},
		cwAngleLimit:            {0x06, 2, rw, true},
		ccwAngleLimit:           {0x08, 2, rw, true},
		highestLimitTemperature: {0x0b, 1, rw, true},
		lowestLimitVoltage:      {0x0c, 1, rw, true},
		highestLimitVoltage:     {0x0d, 1, rw, true},
		maxTorque:               {0x0e, 2, rw, true},
		statusReturnLevel:       {0x10, 1, rw, true},
		alarmLed:                {0x11, 1, rw, true},
		alarmShutdown:           {0x12, 1, rw, true},
		torqueEnable:            {0x18, 1, rw, true},
		led:                     {0x19, 1, rw, true},
		cwComplianceMargin:      {0x1a, 1, rw, true},
		ccwComplianceMargin:     {0x1b, 1, rw, true},
		cwComplianceSlope:       {0x1c, 1, rw, true},
		ccwComplianceSlope:      {0x1d, 1, rw, true},
		goalPosition:            {0x1e, 2, rw, true},
		movingSpeed:             {0x20, 2, rw, true},
		torqueLimit:             {0x22, 2, rw, true},
		presentPosition:         {0x24, 2, ro, true},
		presentSpeed:            {0x26, 2, ro, true},
		presentLoad:             {0x28, 2, ro, true},
		presentVoltage:          {0x2a, 1, ro, true},
		presentTemperature:      {0x2b, 1, ro, true},
		registered:              {0x2c, 1, ro, true},
		moving:                  {0x2e, 1, ro, true},
		lock:                    {0x2f, 1, rw, true},
		punch:                   {0x30, 2, rw, true},
	}
}
