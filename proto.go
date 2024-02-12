package pcc

import (
	"encoding/binary"
	"errors"
	"io"
	"slices"
)

const (
	entLen     = 24
	regLen     = 24
	seqLen     = 24
	procHdrLen = entLen + 4 + 8 + 4 + 4 + 2
)

type versionbyte byte

func (v versionbyte) Major() uint8 {
	return uint8(v >> 4)
}

func (v versionbyte) Minor() uint8 {
	return uint8(v & 0xf)
}

type pktType uint8

const (
	_ pktType = iota // Fobidden zero packet.
	pktSetConfig
	pktDoProcess
)

// header is the header of a pcc packet. Below is the layout of the header, bit lengths in parentheses. Reserved bits must be zero for first version.
//
//	| Version(8) | RSV1(16) | Packet Type(8) | RSV2(16) | Packet Length(16) | Packet ID(16) |
type header struct {
	// protoversion is the protocol protoversion.
	protoversion uint8
	// rsv1 uint16

	ptype pktType
	// rsv2 uint16

	plen uint16
	pid  uint16
}

type Rx struct {
	lastHeader header
	lastEnt    entity
	r          io.ReadCloser
	buf        [24]byte
	setConfig  ControllerConfig
	proc       Process
}

func (r *Rx) Reset(rd io.ReadCloser) {
	r.close()
	r.r = rd
	r.lastHeader = header{}
	r.lastEnt = entity{}
}

func (r *Rx) ReceiveNext() (int, error) {
	if r.r == nil {
		return 0, io.EOF
	}
	n, err := r.recv()
	if err != nil && n > 0 {
		r.close() // On error and non-zero bytes read we've desynced, close connection.
	}
	return n, err
}

func (r *Rx) recv() (n int, err error) {
	// nbr (number of bytes read) is a temporary variable.
	var nbr int
	r.buf[0] = 0
	nbr, err = r.r.Read(r.buf[:1])
	n += nbr
	if n == 0 {
		goto BADREAD
	}
	if r.buf[0] != 1 {
		err = errors.New("unsupported version or garbage message")
		goto BADREAD
	}
	nbr, err = r.r.Read(r.buf[1:10])
	n += nbr
	if n != 7 {
		goto BADREAD
	}
	// Check header RSV1 and RSV2.
	if binary.BigEndian.Uint16(r.buf[1:3]) != 0 || binary.BigEndian.Uint16(r.buf[4:6]) != 0 {
		err = errors.New("reserved bytes are not zero")
		goto BADREAD
	}
	r.lastHeader = header{
		protoversion: r.buf[0],
		ptype:        pktType(r.buf[1]),
		plen:         binary.BigEndian.Uint16(r.buf[6:8]),
		pid:          binary.BigEndian.Uint16(r.buf[8:10]),
	}

	if r.lastHeader.ptype.HasLeadingEntityHeader() {
		// If packet contains a common entity header decode it early.
		if r.lastHeader.plen < 8 {
			err = errors.New("packet length too short")
			goto BADREAD
		}
		nbr, err = r.r.Read(r.buf[:entLen])
		n += nbr
		if nbr != entLen {
			goto BADREAD
		}
		r.lastEnt = decodeEntity(r.buf[:entLen])
	}

	switch r.lastHeader.ptype {
	case pktDoProcess:
		/*

		 DO PROCESS.

		*/

		// | CRC(32) | Process Flags(64) | Start(32) | ConfigCRC(32) | nUnits(16) |
		const procHdrLen = procHdrLen - entLen
		nbr, err = r.r.Read(r.buf[:procHdrLen])
		n += nbr
		if nbr != procHdrLen {
			goto BADREAD
		}
		r.proc.entity = r.lastEnt
		r.proc.CRC = binary.BigEndian.Uint32(r.buf[:4])
		r.proc.Flags = ProcessFlags(binary.BigEndian.Uint64(r.buf[4:12]))
		r.proc.Start = maybetime(binary.BigEndian.Uint32(r.buf[12:16]))
		r.proc.ConfigCRC = binary.BigEndian.Uint32(r.buf[16:20])
		nUnits := int16(binary.BigEndian.Uint16(r.buf[20:22]))
		if nUnits <= 0 {
			err = errors.New("zero units in Process")
			goto BADREAD
		}
		r.proc.Units = r.proc.Units[:0]
		for i := int16(0); i < nUnits; i++ {
			// | Unit Flags(8) | Num Forks(8) | Sequence(16) | NextUnit(16) | Forks(16*NumForks) |
			nbr, err = r.r.Read(r.buf[:6])
			n += nbr
			if nbr != 8 {
				goto BADREAD
			}
			nForks := int8(r.buf[1])
			if nForks <= 0 {
				err = errors.New("invalid number of forks")
				goto BADREAD
			}

			u := extend1(&r.proc.Units)
			u.Flags = UnitFlags(r.buf[0])
			u.Sequence = binary.BigEndian.Uint16(r.buf[2:4])
			u.Next = binary.BigEndian.Uint16(r.buf[4:6])
			u.Forks = u.Forks[:0]
			for j := int8(0); j < nForks; j++ {
				nbr, err = r.r.Read(r.buf[:2])
				n += nbr
				if nbr != 2 {
					goto BADREAD
				}
				u.Forks = append(u.Forks, binary.BigEndian.Uint16(r.buf[0:]))
			}
		}

	case pktSetConfig:
		/*

		 SET CONFIGURATION.

		*/

		// | CRC(32) | nRegisters(8) | nSequences(8) | nCommandLists(8) | rsv(8) |
		nbr, err = r.r.Read(r.buf[:8])
		n += nbr
		if nbr != 8 {
			goto BADREAD
		}
		// gotCRC := binary.BigEndian.Uint32(r.buf[:4])
		// We cap the maximum length of steps,regs,seqs to 127 to allow for
		// future expansion of the protocol with 8th bit.
		nRegisters := int8(r.buf[4])
		nSequences := int8(r.buf[5])
		nCommandLists := int8(r.buf[6])
		rsv := r.buf[7]
		if rsv != 0 || nCommandLists <= 0 || nRegisters < 0 || nSequences <= 0 {
			err = errors.New("invalid setconfig header")
			goto BADREAD
		}

		r.setConfig = ControllerConfig{entity: r.lastEnt} // Reset config.
		for i := int8(0); i < nRegisters; i++ {
			// | Name(64) | Base(8) | rsv(8) | Dimension(32) | Value(64) |
			nbr, err = r.r.Read(r.buf[:regLen])
			n += nbr
			if nbr != regLen {
				goto BADREAD
			}

			r.setConfig.Registers = append(r.setConfig.Registers, decodeRegister(r.buf[:regLen]))
		}

		for i := int8(0); i < nSequences; i++ {
			// | Name(64) | nSteps(8) |
			nbr, err = r.r.Read(r.buf[:8])
			n += nbr
			if nbr != seqLen {
				goto BADREAD
			}

			seq := extend1(&r.setConfig.Sequences)
			copy(seq.Name[:], r.buf[:8])
			nsteps := int8(r.buf[8])
			if nsteps <= 0 {
				err = errors.New("invalid sequence header")
				goto BADREAD
			}
			for j := int8(0); j < nsteps; j++ {
				// | CommandIdx(16) | ModuleID(16) |
				nbr, err = r.r.Read(r.buf[:4])
				n += nbr
				if nbr != 4 {
					goto BADREAD
				}
				seq.Command = append(seq.Command, decodeSequenceCmd(r.buf[:4]))
			}
		}

		for i := int8(0); i < nCommandLists; i++ {
			// | APIVersion(8) | rsv(8) | ModuleType(16) | nCmds(16) |
			const steplistHdrLen = 6
			nbr, err = r.r.Read(r.buf[:steplistHdrLen])
			n += nbr
			if n != steplistHdrLen {
				goto BADREAD
			}
			sl := extend1(&r.setConfig.CommandLists)
			sl.APIVersion = uint8(r.buf[0])
			sl.ModuleType = moduletype(r.buf[2])
			sl.ModuleType = binary.BigEndian.Uint16(r.buf[3:5])
			nCmds := int16(binary.BigEndian.Uint16(r.buf[5:7]))
			if nCmds <= 0 || nCmds > 127 {
				err = errors.New("invalid commandlist header")
				goto BADREAD
			}
			for j := int16(0); j < nCmds; j++ {
				// | Procedure(16) | Arglen(8) | Args(Arglen) |
				nbr, err = r.r.Read(r.buf[:3])
				n += nbr
				if n != 3 {
					goto BADREAD
				}

				cmd := extend1(&sl.Commands)
				cmd.Procedure = procedure(binary.BigEndian.Uint16(r.buf[:2]))
				arglen := r.buf[2]
				cmd.Args = cmd.Args[:0]
				cmd.Args = slices.Grow(cmd.Args, int(arglen))
				cmd.Args = cmd.Args[:arglen]
				nbr, err = r.r.Read(cmd.Args[:arglen])
				n += nbr
				if n != int(arglen) {
					goto BADREAD
				}
			}
		}

	}
	return 0, nil

BADREAD:
	if err == nil {
		err = io.ErrNoProgress
	}
	return n, err
}

func (r *Rx) close() {
	if r.r != nil {
		r.r.Close()
		r.r = nil
	}
}

func (r *Rx) decodeConfig() {
	// Read CRC+configlens

}

func decodeEntity(b []byte) (ent entity) {
	_ = b[entLen-1]
	binary.BigEndian.Uint64(b[:8])
	copy(ent.uuid[:], b[8:24])
	return ent
}

func decodeRegister(b []byte) (reg Register) {
	_ = b[regLen-1]
	copy(reg.Name[:], b[:8])
	reg.Base = int8(b[8])
	// Skip over 4 reserved bytes.
	copy(reg.Dimension.dims[:], b[12:16])
	reg.Value = int64(binary.BigEndian.Uint64(b[16:24]))
	return reg
}

func decodeSequenceCmd(b []byte) (seq [2]int) {
	seq[0] = int(binary.BigEndian.Uint16(b[:2]))
	seq[1] = int(binary.BigEndian.Uint16(b[2:4]))
	return seq
}

func (ptype pktType) HasLeadingEntityHeader() bool {
	return ptype == pktDoProcess || ptype == pktSetConfig
}

func (cfg ControllerConfig) Reset() {
	cfg.Registers = cfg.Registers[:0]
	cfg.Sequences = cfg.Sequences[:0]
	cfg.CommandLists = cfg.CommandLists[:0]
	cfg.CRC = 0
	cfg.entity = entity{}
}

func extend1[T any](b *[]T) *T {
	var t T
	if cap(*b) == len(*b) {
		*b = append(*b, t) // Need to grow the slice to append.
	} else {
		*b = (*b)[:len(*b)+1]
	}
	return &(*b)[len(*b)-1]

}
