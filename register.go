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

	// The range of values which this register can be set to. This only applies
	// is the register is RW.
	min int
	max int
}

// See: http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm
var registers map[RegName]*Register

func init() {
	registers = map[RegName]*Register{
		modelNumber:             {0x00, 2, ro, true, 0, 0},
		firmwareVersion:         {0x02, 1, ro, true, 0, 0},
		servoID:                 {0x03, 1, rw, true, 0, 1024}, // renamed from ID for clarity
		baudRate:                {0x04, 1, rw, true, 0, 1024},
		returnDelayTime:         {0x05, 1, rw, true, 0, 1024},
		cwAngleLimit:            {0x06, 2, rw, true, 0, 1024},
		ccwAngleLimit:           {0x08, 2, rw, true, 0, 1024},
		highestLimitTemperature: {0x0b, 1, rw, true, 0, 1024},
		lowestLimitVoltage:      {0x0c, 1, rw, true, 0, 1024},
		highestLimitVoltage:     {0x0d, 1, rw, true, 0, 1024},
		maxTorque:               {0x0e, 2, rw, true, 0, 1024},
		statusReturnLevel:       {0x10, 1, rw, true, 0, 1024},
		alarmLed:                {0x11, 1, rw, true, 0, 1024},
		alarmShutdown:           {0x12, 1, rw, true, 0, 1024},
		torqueEnable:            {0x18, 1, rw, true, 0, 1024},
		led:                     {0x19, 1, rw, true, 0, 1024},
		cwComplianceMargin:      {0x1a, 1, rw, true, 0, 1024},
		ccwComplianceMargin:     {0x1b, 1, rw, true, 0, 1024},
		cwComplianceSlope:       {0x1c, 1, rw, true, 0, 1024},
		ccwComplianceSlope:      {0x1d, 1, rw, true, 0, 1024},
		goalPosition:            {0x1e, 2, rw, true, 0, 1024},
		movingSpeed:             {0x20, 2, rw, true, 0, 1024},
		torqueLimit:             {0x22, 2, rw, true, 0, 1024},
		presentPosition:         {0x24, 2, ro, false, 0, 0},
		presentSpeed:            {0x26, 2, ro, true, 0, 0},
		presentLoad:             {0x28, 2, ro, true, 0, 0},
		presentVoltage:          {0x2a, 1, ro, false, 0, 0},
		presentTemperature:      {0x2b, 1, ro, true, 0, 0},
		registered:              {0x2c, 1, ro, true, 0, 0},
		moving:                  {0x2e, 1, ro, true, 0, 0},
		lock:                    {0x2f, 1, rw, true, 0, 1024},
		punch:                   {0x30, 2, rw, true, 0, 1024},
	}
}
