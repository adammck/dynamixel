package dynamixel

type Access int

const(
	RO Access = iota
	RW
)

type Register struct {
	address byte
	length int
	access Access
}
