package registers

//go:generate stringer -type=RegName
type RegName int
type Access int
type Map map[RegName]*Register

type Register struct {
	Address   byte
	Length    int
	Access    Access
	Cacheable bool

	// The range of values which this register can be set to. This only applies
	// is the register is RW.
	Min int
	Max int
}

const (

	// Register names are used to refer to a specific value in a servo's EEPROM
	// or RAM. The address (and presence) of each varies between servo models.
	ModelNumber RegName = iota
	FirmwareVersion
	ServoID
	BaudRate
	ReturnDelayTime
	CwAngleLimit
	CcwAngleLimit
	HighestLimitTemperature
	LowestLimitVoltage
	HighestLimitVoltage
	MaxTorque
	StatusReturnLevel
	AlarmLed
	AlarmShutdown
	TorqueEnable
	Led
	CwComplianceMargin
	CcwComplianceMargin
	CwComplianceSlope
	CcwComplianceSlope
	GoalPosition
	MovingSpeed
	TorqueLimit
	PresentPosition
	PresentSpeed
	PresentLoad
	PresentVoltage
	PresentTemperature
	Registered
	Moving
	Lock
	Punch

	// Access Levels specify whether a register is hard-coded into the servo
	// (e.g. the model number), or is a value which can be changed (e.g. the
	// identity).
	RO Access = iota
	RW
)
