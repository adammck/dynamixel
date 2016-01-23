package servo

import (
	"errors"
	"fmt"

	"github.com/adammck/dynamixel/utils"
)

//
// -- Registers
//
//    These methods are getters for the various registers in the control table.
//    Some of them (where register.cacheable == true) just read from the cache,
//    while others read the actual control table every time.
//
// TODO: Each of the following registers should have a corresponding reader, and
//       the R/W registers (marked with an asterisk) should have a writer. They
//       should all receive and return ints or bools, rather than bytes.
//
// TODO: These methods should probably be generated from the list of registers,
//       especially if/when we support multiple models with different sets.
//
// modelNumber
// firmwareVersion
// servoID*
// baudRate*
// returnDelayTime*
// cwAngleLimit*
// ccwAngleLimit*
// highestLimitTemperature*
// lowestLimitVoltage*
// highestLimitVoltage*
// maxTorque*
// statusReturnLevel*
// alarmLed*
// alarmShutdown*
// torqueEnable*
// led*
// cwComplianceMargin*
// ccwComplianceMargin*
// cwComplianceSlope*
// ccwComplianceSlope*
// goalPosition*
// movingSpeed*
// torqueLimit*
// presentPosition
// presentSpeed
// presentLoad
// presentVoltage
// presentTemperature
// registered
// moving
// lock*
// punch*
//

func (servo *Servo) ModelNumber() (int, error) {
	return servo.getRegister(*registers[modelNumber])
}

func (servo *Servo) FirmwareVersion() (int, error) {
	return servo.getRegister(*registers[firmwareVersion])
}

func (servo *Servo) ServoID() (int, error) {
	return servo.getRegister(*registers[servoID])
}

// SetServoID changes the identity of the servo.
// This is stored in EEPROM, so will persist between reboots.
func (servo *Servo) SetServoID(ident int) error {
	return servo.setRegister(*registers[servoID], ident)
}

func (servo *Servo) BaudRate() (int, error) {
	return servo.getRegister(*registers[baudRate])
}

func (servo *Servo) SetBaudRate(v int) error {
	return servo.setRegister(*registers[baudRate], v)
}

func (servo *Servo) ReturnDelayTime() (int, error) {
	return servo.getRegister(*registers[returnDelayTime])
}

func (servo *Servo) SetReturnDelayTime(v int) error {
	return servo.setRegister(*registers[returnDelayTime], v)
}

func (servo *Servo) CWAngleLimit() (int, error) {
	return servo.getRegister(*registers[cwAngleLimit])
}

func (servo *Servo) SetCWAngleLimit(v int) error {
	return servo.setRegister(*registers[cwAngleLimit], v)
}

func (servo *Servo) CCWAngleLimit() (int, error) {
	return servo.getRegister(*registers[ccwAngleLimit])
}

func (servo *Servo) SetCCWAngleLimit(v int) error {
	return servo.setRegister(*registers[ccwAngleLimit], v)
}

func (servo *Servo) HighestLimitTemperature() (int, error) {
	return servo.getRegister(*registers[highestLimitTemperature])
}

func (servo *Servo) SetHighestLimitTemperature(v int) error {
	return servo.setRegister(*registers[highestLimitTemperature], v)
}

func (servo *Servo) LowestLimitVoltage() (int, error) {
	return servo.getRegister(*registers[lowestLimitVoltage])
}

func (servo *Servo) SetLowestLimitVoltage(v int) error {
	return servo.setRegister(*registers[lowestLimitVoltage], v)
}

func (servo *Servo) HighestLimitVoltage() (int, error) {
	return servo.getRegister(*registers[highestLimitVoltage])
}

func (servo *Servo) SetHighestLimitVoltage(v int) error {
	return servo.setRegister(*registers[highestLimitVoltage], v)
}

func (servo *Servo) MaxTorque() (int, error) {
	return servo.getRegister(*registers[maxTorque])
}

func (servo *Servo) SetMaxTorque(v int) error {
	return servo.setRegister(*registers[maxTorque], v)
}

func (servo *Servo) StatusReturnLevel() (int, error) {
	return servo.getRegister(*registers[statusReturnLevel])
}

// Sets the status return level. Possible values are:
//
// 0 = Only respond to PING commands
// 1 = Only respond to PING and READ commands
// 2 = Respond to all commands
//
// Servos default to 2, but retain the value so long as they're powered up. This
// makes it a very good idea to explicitly set the value after connecting, to
// avoid waiting for status packets which will never arrive.
//
// See: dxl_ax_actuator.htm#Actuator_Address_10
func (servo *Servo) SetStatusReturnLevel(value int) error {
	reg := *registers[statusReturnLevel]

	if value < reg.Min || value > reg.Max {
		return fmt.Errorf("invalid Status Return Level value: %d", value)
	}

	ident, err := servo.ServoID()
	if err != nil {
		return err
	}

	// Call Network.WriteData directly, rather than via servo.writeData, because
	// the return status level will depend upon the new level, rather than the
	// current level cache. We don't want to update that until we're sure that
	// the write was successful.
	err = servo.Network.WriteData(uint8(ident), (value == 2), reg.Address, utils.Low(value))
	if err != nil {
		return err
	}

	return nil
}

func (servo *Servo) AlarmLED() (int, error) {
	return servo.getRegister(*registers[alarmLed])
}

func (servo *Servo) SetAlarmLED(v int) error {
	return servo.setRegister(*registers[alarmLed], v)
}

func (servo *Servo) AlarmShutdown() (int, error) {
	return servo.getRegister(*registers[alarmShutdown])
}

func (servo *Servo) SetAlarmShutdown(v int) error {
	return servo.setRegister(*registers[alarmShutdown], v)
}

func (servo *Servo) TorqueEnable() (bool, error) {
	v, err := servo.getRegister(*registers[torqueEnable])
	return utils.IntToBool(v), err
}

// SetTorqueEnable enables or disables torque.
func (servo *Servo) SetTorqueEnable(state bool) error {
	return servo.setRegister(*registers[torqueEnable], utils.BoolToInt(state))
}

// LED returns the current state of the servo's LED.
// TODO: Should we continue to return bool here, or expose the int?
func (servo *Servo) LED() (bool, error) {
	v, err := servo.getRegister(*registers[led])
	return utils.IntToBool(v), err
}

// Enables or disables the servo's LED.
func (servo *Servo) SetLED(state bool) error {
	return servo.setRegister(*registers[led], utils.BoolToInt(state))
}

func (servo *Servo) CWComplianceMargin() (int, error) {
	return servo.getRegister(*registers[cwComplianceMargin])
}

func (servo *Servo) SetCWComplianceMargin(v int) error {
	return servo.setRegister(*registers[cwComplianceMargin], v)
}

func (servo *Servo) CCWComplianceMargin() (int, error) {
	return servo.getRegister(*registers[ccwComplianceMargin])
}

func (servo *Servo) SetCCWComplianceMarginval(v int) error {
	return servo.setRegister(*registers[ccwComplianceMargin], v)
}

func (servo *Servo) CWComplianceSlope() (int, error) {
	return servo.getRegister(*registers[cwComplianceSlope])
}

func (servo *Servo) SetCWComplianceSlope(v int) error {
	return servo.setRegister(*registers[cwComplianceSlope], v)
}

func (servo *Servo) CCWComplianceSlope() (int, error) {
	return servo.getRegister(*registers[ccwComplianceSlope])
}

func (servo *Servo) SetCCWComplianceSlope(v int) error {
	return servo.setRegister(*registers[ccwComplianceSlope], v)
}

func (servo *Servo) GoalPosition() (int, error) {
	return servo.getRegister(*registers[goalPosition])
}

// SetGoalPosition sets the goal position.
//
// TODO: Reject if the servo is in wheel mode (where CW and CCW angle limit
//       is zero).
//
func (servo *Servo) SetGoalPosition(pos int) error {
	return servo.setRegister(*registers[goalPosition], pos)
}

// MovingSpeed returns the current moving speed. This is not the speed at which
// the motor is moving, it's the speed at which the servo wants to move.
func (servo *Servo) MovingSpeed() (int, error) {
	return servo.getRegister(*registers[movingSpeed])
}

// SetMovingSpeed the moving speed.
func (servo *Servo) SetMovingSpeed(speed int) error {
	return servo.setRegister(*registers[movingSpeed], speed)
}

func (servo *Servo) TorqueLimit() (int, error) {
	return servo.getRegister(*registers[torqueLimit])
}

func (servo *Servo) SetTorqueLimit(val int) error {
	return servo.setRegister(*registers[torqueLimit], val)
}

func (servo *Servo) PresentPosition() (int, error) {
	return servo.getRegister(*registers[presentPosition])
}

func (servo *Servo) PresentSpeed() (int, error) {
	return servo.getRegister(*registers[presentSpeed])
}

func (servo *Servo) PresentVoltage() (int, error) {
	return servo.getRegister(*registers[presentVoltage])
}

func (servo *Servo) PresentLoad() (int, error) {
	return servo.getRegister(*registers[presentLoad])
}

func (servo *Servo) PresentTemperature() (int, error) {
	return servo.getRegister(*registers[presentTemperature])
}

func (servo *Servo) Registered() (int, error) {
	return servo.getRegister(*registers[registered])
}

func (servo *Servo) Moving() (int, error) {
	return servo.getRegister(*registers[moving])
}

// TODO: Rename this to avoid confusion?
func (servo *Servo) Lock() (int, error) {
	return servo.getRegister(*registers[lock])
}

func (servo *Servo) SetLock(isLocked int) error {
	reg := *registers[lock]

	// Can't unlock when servo is locked, so if we know that's the case, don't
	// bother trying. Can be overriden by clearing the cache.
	//
	// TODO: Add a method to read ints from the cache. If we used getRegister,
	//       we risk accidentally (in the case of a bug) reading from the actual
	//       device, which would be slow and weird.

	if isLocked == 0 && servo.cache[reg.Address] == byte(1) {
		return errors.New("EEPROM can't be unlocked; must be power-cycled")
	}

	return servo.setRegister(reg, isLocked)
}
