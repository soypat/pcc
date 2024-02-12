package pcc

// Process definition types.
type (
	uuid_t     = [16]byte
	procedure  = uint16
	moduletype = uint16
	maybetime  = uint32
)

// NewDimension creates a new dimension from the given exponents.
func NewDimension(len, mass, time, temp, icurrent, luminosity, amount, _ int8) Dimension {
	return Dimension{
		dims: [4]byte{
			i8ToI4(len) | i8ToI4(mass)<<4,
			i8ToI4(time) | i8ToI4(temp)<<4,
			i8ToI4(icurrent) | i8ToI4(luminosity)<<4,
			i8ToI4(amount) | i8ToI4(0)<<4,
		},
	}
}

// Dimension represents the dimensions of a physical quantity.
type Dimension struct {
	// dims contains 7 int4's representing the exponent of primitive dimensions:
	//  0. Distance dimension (L)
	//  1. Mass dimension (M)
	//  2. Time dimension (T)
	//  3. Temperature dimension (K)
	//  4. Electric current dimension (I)
	//  5. Luminous intensity dimension (J)
	//  6. Amount or quantity dimension (N). i.e: moles, particles, electric pulses, etc.
	// Since these are int4s they are in the range of -8 to 7. Last 4 bits of fourth byte are unused.
	dims [4]byte
}

func (d Dimension) ExpLength() int8      { return i4ToI8(d.dims[0] & 0xf) }
func (d Dimension) ExpMass() int8        { return i4ToI8(d.dims[0] >> 4) }
func (d Dimension) ExpTime() int8        { return i4ToI8(d.dims[1] & 0xf) }
func (d Dimension) ExpTemperature() int8 { return i4ToI8(d.dims[1] >> 4) }
func (d Dimension) ExpCurrent() int8     { return i4ToI8(d.dims[2] & 0xf) }
func (d Dimension) ExpLuminous() int8    { return i4ToI8(d.dims[2] >> 4) }
func (d Dimension) ExpAmount() int8      { return i4ToI8(d.dims[3] & 0xf) }

func i8ToI4(c int8) byte {
	if c < 0 {
		c |= 1 << 3
	}
	return byte(c) & 0xf
}

// i4ToI8 converts lower 4 bits of a byte to a signed 4 bit integer.
func i4ToI8(c byte) int8 {
	c &= 0xf
	if c&0x8 != 0 {
		c = c | 0xf0
	}
	return int8(c)
}

// entity is a base type for process control flow data types that
// define the actuations that occured in a process.
//
// If a type contains an entity, it is a good sign that anytime the type
// is exchanged with a process controller, it should be logged/saved to a database.
type entity struct {
	flags uint64
	// uuid is the unique identifier for the process so it may be kept track of.
	uuid uuid_t
}

type ProcessFlags uint64

type UnitFlags uint8

// Unit represents a process control flow step in a Process.
// Actions of a Unit are represented as indices corresponding to the
// configuration of the process controller.
type Unit struct {
	Flags    UnitFlags
	Sequence uint16
	Next     uint16
	Forks    []uint16
}

type Register struct {
	Name [8]byte
	Base int8
	// rsv 	 [3]byte
	Dimension Dimension
	Value     int64
}

type Sequence struct {
	// Name is the human readable name of the sequence.
	Name [8]byte
	// Command is a list of tuples containing:
	//  0. The index+1 w.r.t CommandList of the command to be executed.
	//  1. The ID of the module to execute the command on. i.e: pump 2, valve 3, etc.
	Command [][2]int
}

type CommandList struct {
	// APIVersion is the version of the API available for the module.
	APIVersion uint8
	// rsv [1]byte

	// ModuleType is the type of module that this command list is for.
	ModuleType moduletype
	// Commands is a list of commands that may be executed by the module.
	Commands []Command
}

type Command struct {
	Procedure procedure
	Args      []byte
}

// Process is a complete definition of a process to be executed by the process controller.
type Process struct {
	entity
	// CRC is the CRC32 of the process packet, not including the CRC field (zeroed).
	CRC   uint32
	Flags ProcessFlags
	Start maybetime

	// ConfigCRC is the CRC32 of the configuration stored in the process controller. Must match.
	ConfigCRC uint32
	// Units defines the process actions to be performed. The first unit is the start of the process.
	Units []Unit
}

// ControllerConfig is the configuration of the process controller.
type ControllerConfig struct {
	entity
	// CRC is the CRC32 of the configuration packet, not including the CRC field (zeroed).
	CRC uint32
	// Registers contains the static configuration of the process controller.
	Registers []Register
	// CommandLists contains available actions that may be performed
	// by the modules available to the process controller.
	CommandLists []CommandList
	// Registry contains the static configuration of the process controller.

	// Sequences contains
	Sequences []Sequence
}

func (e entity) Version() uint8 {
	return uint8(e.flags & 0xF)
}

func (e entity) UUID() uuid_t {
	return e.uuid
}
