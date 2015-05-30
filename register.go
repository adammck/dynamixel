package dynamixel

type RegName int
type Access int

const (
	ro Access = iota
	rw

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
)

type Register struct {
	address byte
	length  int
	access  Access
}

var registers map[RegName]*Register

func init() {
	registers = map[RegName]*Register{
		modelNumber:             {0x00, 2, ro},
		firmwareVersion:         {0x02, 1, ro},
		servoID:                 {0x03, 1, rw}, // renamed from ID for clarity
		baudRate:                {0x04, 1, rw},
		returnDelayTime:         {0x05, 1, rw},
		cwAngleLimit:            {0x06, 2, rw},
		ccwAngleLimit:           {0x08, 2, rw},
		highestLimitTemperature: {0x0b, 1, rw},
		lowestLimitVoltage:      {0x0c, 1, rw},
		highestLimitVoltage:     {0x0d, 1, rw},
		maxTorque:               {0x0e, 2, rw},
		statusReturnLevel:       {0x10, 1, rw},
		alarmLed:                {0x11, 1, rw},
		alarmShutdown:           {0x12, 1, rw},
		torqueEnable:            {0x18, 1, rw},
		led:                     {0x19, 1, rw},
		cwComplianceMargin:      {0x1a, 1, rw},
		ccwComplianceMargin:     {0x1b, 1, rw},
		cwComplianceSlope:       {0x1c, 1, rw},
		ccwComplianceSlope:      {0x1d, 1, rw},
		goalPosition:            {0x1e, 2, rw},
		movingSpeed:             {0x20, 2, rw},
		torqueLimit:             {0x22, 2, rw},
		presentPosition:         {0x24, 2, ro},
		presentSpeed:            {0x26, 2, ro},
		presentLoad:             {0x28, 2, ro},
		presentVoltage:          {0x2a, 1, ro},
		presentTemperature:      {0x2b, 1, ro},
		registered:              {0x2c, 1, ro},
		moving:                  {0x2e, 1, ro},
		lock:                    {0x2f, 1, rw},
		punch:                   {0x30, 2, rw},
	}
}
