package registers

//go:generate stringer -type=RegName
type RegName int
type Access int
type Map map[RegName]*Register

type Register struct {

	// TODO: Make address a plain int, since proto1 wants a byte, and proto2
	//       wants a uint16. We have to cast it either way.
	Address byte

	Length int
	Access Access

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
	RegisteredInstruction
	Moving
	Lock
	Punch

	ControlMode         // XL-320
	DGain               // XL-320
	IGain               // XL-320
	PGain               // XL-320
	HardwareErrorStatus // XL-320
	GoalVelocity
	GoalTorque

	// Access Levels specify whether a register is hard-coded into the servo
	// (e.g. the model number), or is a value which can be changed (e.g. the
	// identity). The zero-value is RO.
	RO Access = 0
	RW Access = 1
)
