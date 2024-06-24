package pcc

import "github.com/soypat/si"

// Process definition types.
type (
	uuid_t     = [16]byte
	procedure  = uint16
	moduletype = uint16
	maybetime  = uint32
)

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

// UnitFlags decide the behavior of a Unit.
// | On End Bits(2) | RSV(6) |
type UnitFlags uint8

const (
	// if any of these bits are set, the unit is an end unit (finishes process).
	unitFlagEndMask = 0b11
)

func (f UnitFlags) IsEnd() bool {
	return f&unitFlagEndMask == 1
}

func (f UnitFlags) IsRestart() bool {
	return f&unitFlagEndMask == 2
}

// Unit represents a process control flow step in a Process.
// Actions of a Unit are represented as indices corresponding to the
// configuration of the process controller.
type Unit struct {
	Flags UnitFlags
	// Sequence is the index of this unit's sequence in the process controller's configuration.
	Sequence uint16
	// Next is the index of the next Unit in the process.
	Next uint16
	// Forks are the index of the forks to be started after running the unit.
	Forks []uint16
}

type Register struct {
	Name [8]byte
	Base int8
	// rsv 	 [3]byte
	Dimension si.Dimension
	Value     int64
}

/*
// AgnosticSequence is a template sequence whose modules are generically defined on execution.
type AgnosticSequence struct {
	CtlModules []struct {
		Purpose    []byte
		ModuleType moduletype
		APIVersion uint8
	}
	Commands []AgnosticSequenceCommand
}

type AgnosticSequenceCommand struct {
	CtlModuleIdx uint16
	CommandIdx   uint16
}
*/

type Sequence struct {
	// Name is the human readable name of the sequence.
	Name [8]byte
	// Commands lists the commands to be executed in the sequence, in order.
	Commands []SequenceCommand
}

type SequenceCommand struct {
	// CommandIndex is the 0-based index of the Command in the CommandList.
	CommandIndex uint16
	APIVersion   uint8
	// The identifier of the module to execute the command on.
	// For example, there may be several pumps in a process controller, each with a unique identifier.
	ModuleID uint8
	// ModuleType is the type of the module to execute the command on. Together with
	// APIVersion they select the CommandList to use.
	ModuleType moduletype
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
