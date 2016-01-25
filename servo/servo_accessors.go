package servo

import (
	"errors"
	"fmt"

	reg "github.com/adammck/dynamixel/registers"
	"github.com/adammck/dynamixel/utils"
)

// These methods are getters for the various registers in the control table.
// Some of them (where register.cacheable == true) just read from the cache,
// while others read the actual control table every time.
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

func (s *Servo) ModelNumber() (int, error) {
	return s.getRegister(reg.ModelNumber)
}

func (s *Servo) FirmwareVersion() (int, error) {
	return s.getRegister(reg.FirmwareVersion)
}

func (s *Servo) ServoID() (int, error) {
	return s.getRegister(reg.ServoID)
}

// SetServoID changes the identity of the servo.
// This is stored in EEPROM, so will persist between reboots.
func (s *Servo) SetServoID(ident int) error {
	return s.setRegister(reg.ServoID, ident)
}

func (s *Servo) BaudRate() (int, error) {
	return s.getRegister(reg.BaudRate)
}

func (s *Servo) SetBaudRate(v int) error {
	return s.setRegister(reg.BaudRate, v)
}

func (s *Servo) ReturnDelayTime() (int, error) {
	return s.getRegister(reg.ReturnDelayTime)
}

func (s *Servo) SetReturnDelayTime(v int) error {
	return s.setRegister(reg.ReturnDelayTime, v)
}

func (s *Servo) CWAngleLimit() (int, error) {
	return s.getRegister(reg.CwAngleLimit)
}

func (s *Servo) SetCWAngleLimit(v int) error {
	return s.setRegister(reg.CwAngleLimit, v)
}

func (s *Servo) CCWAngleLimit() (int, error) {
	return s.getRegister(reg.CcwAngleLimit)
}

func (s *Servo) SetCCWAngleLimit(v int) error {
	return s.setRegister(reg.CcwAngleLimit, v)
}

func (s *Servo) HighestLimitTemperature() (int, error) {
	return s.getRegister(reg.HighestLimitTemperature)
}

func (s *Servo) SetHighestLimitTemperature(v int) error {
	return s.setRegister(reg.HighestLimitTemperature, v)
}

func (s *Servo) LowestLimitVoltage() (int, error) {
	return s.getRegister(reg.LowestLimitVoltage)
}

func (s *Servo) SetLowestLimitVoltage(v int) error {
	return s.setRegister(reg.LowestLimitVoltage, v)
}

func (s *Servo) HighestLimitVoltage() (int, error) {
	return s.getRegister(reg.HighestLimitVoltage)
}

func (s *Servo) SetHighestLimitVoltage(v int) error {
	return s.setRegister(reg.HighestLimitVoltage, v)
}

func (s *Servo) MaxTorque() (int, error) {
	return s.getRegister(reg.MaxTorque)
}

func (s *Servo) SetMaxTorque(v int) error {
	return s.setRegister(reg.MaxTorque, v)
}

func (s *Servo) StatusReturnLevel() (int, error) {
	return s.getRegister(reg.StatusReturnLevel)
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
func (s *Servo) SetStatusReturnLevel(value int) error {
	reg := s.registers[reg.StatusReturnLevel]

	if value < reg.Min || value > reg.Max {
		return fmt.Errorf("invalid Status Return Level value: %d", value)
	}

	ident, err := s.ServoID()
	if err != nil {
		return err
	}

	// Call Network.WriteData directly, rather than via writeData, because the
	// return status level will depend upon the new level, rather than the
	// current level cache. We don't want to update that until we're sure that
	// the write was successful.
	err = s.Network.WriteData(uint8(ident), (value == 2), reg.Address, utils.Low(value))
	if err != nil {
		return err
	}

	return nil
}

func (s *Servo) AlarmLED() (int, error) {
	return s.getRegister(reg.AlarmLed)
}

func (s *Servo) SetAlarmLED(v int) error {
	return s.setRegister(reg.AlarmLed, v)
}

func (s *Servo) AlarmShutdown() (int, error) {
	return s.getRegister(reg.AlarmShutdown)
}

func (s *Servo) SetAlarmShutdown(v int) error {
	return s.setRegister(reg.AlarmShutdown, v)
}

func (s *Servo) TorqueEnable() (bool, error) {
	v, err := s.getRegister(reg.TorqueEnable)
	return utils.IntToBool(v), err
}

// SetTorqueEnable enables or disables torque.
func (s *Servo) SetTorqueEnable(state bool) error {
	return s.setRegister(reg.TorqueEnable, utils.BoolToInt(state))
}

// LED returns the current state of the servo's LED.
// TODO: Should we continue to return bool here, or expose the int?
func (s *Servo) LED() (bool, error) {
	v, err := s.getRegister(reg.Led)
	return utils.IntToBool(v), err
}

// Enables or disables the servo's LED.
func (s *Servo) SetLED(state bool) error {
	return s.setRegister(reg.Led, utils.BoolToInt(state))
}

func (s *Servo) CWComplianceMargin() (int, error) {
	return s.getRegister(reg.CwComplianceMargin)
}

func (s *Servo) SetCWComplianceMargin(v int) error {
	return s.setRegister(reg.CwComplianceMargin, v)
}

func (s *Servo) CCWComplianceMargin() (int, error) {
	return s.getRegister(reg.CcwComplianceMargin)
}

func (s *Servo) SetCCWComplianceMarginval(v int) error {
	return s.setRegister(reg.CcwComplianceMargin, v)
}

func (s *Servo) CWComplianceSlope() (int, error) {
	return s.getRegister(reg.CwComplianceSlope)
}

func (s *Servo) SetCWComplianceSlope(v int) error {
	return s.setRegister(reg.CwComplianceSlope, v)
}

func (s *Servo) CCWComplianceSlope() (int, error) {
	return s.getRegister(reg.CcwComplianceSlope)
}

func (s *Servo) SetCCWComplianceSlope(v int) error {
	return s.setRegister(reg.CcwComplianceSlope, v)
}

func (s *Servo) GoalPosition() (int, error) {
	return s.getRegister(reg.GoalPosition)
}

// SetGoalPosition sets the goal position.
//
// TODO: Reject if the servo is in wheel mode (where CW and CCW angle limit
//       is zero).
//
func (s *Servo) SetGoalPosition(pos int) error {
	return s.setRegister(reg.GoalPosition, pos)
}

// MovingSpeed returns the current moving speed. This is not the speed at which
// the motor is moving, it's the speed at which the servo wants to move.
func (s *Servo) MovingSpeed() (int, error) {
	return s.getRegister(reg.MovingSpeed)
}

// SetMovingSpeed the moving speed.
func (s *Servo) SetMovingSpeed(speed int) error {
	return s.setRegister(reg.MovingSpeed, speed)
}

func (s *Servo) TorqueLimit() (int, error) {
	return s.getRegister(reg.TorqueLimit)
}

func (s *Servo) SetTorqueLimit(val int) error {
	return s.setRegister(reg.TorqueLimit, val)
}

func (s *Servo) PresentPosition() (int, error) {
	return s.getRegister(reg.PresentPosition)
}

func (s *Servo) PresentSpeed() (int, error) {
	return s.getRegister(reg.PresentSpeed)
}

func (s *Servo) PresentVoltage() (int, error) {
	return s.getRegister(reg.PresentVoltage)
}

func (s *Servo) PresentLoad() (int, error) {
	return s.getRegister(reg.PresentLoad)
}

func (s *Servo) PresentTemperature() (int, error) {
	return s.getRegister(reg.PresentTemperature)
}

func (s *Servo) Registered() (int, error) {
	return s.getRegister(reg.Registered)
}

func (s *Servo) Moving() (int, error) {
	return s.getRegister(reg.Moving)
}

// TODO: Rename this to avoid confusion?
func (s *Servo) Lock() (int, error) {
	return s.getRegister(reg.Lock)
}

func (s *Servo) SetLock(isLocked int) error {
	r := s.registers[reg.Lock]

	// Can't unlock when servo is locked, so if we know that's the case, don't
	// bother trying. Can be overriden by clearing the cache.
	//
	// TODO: Add a method to read ints from the cache. If we used getRegister,
	//       we risk accidentally (in the case of a bug) reading from the actual
	//       device, which would be slow and weird.

	if isLocked == 0 && s.cache[r.Address] == byte(1) {
		return errors.New("EEPROM can't be unlocked; must be power-cycled")
	}

	return s.setRegister(reg.Lock, isLocked)
}
