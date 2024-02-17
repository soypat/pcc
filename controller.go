package pcc

import (
	"errors"
	"fmt"
)

type Controller struct {
	cfg        ControllerConfig
	procedures map[uint32]func(proc procedure, id uint8, arg []byte) error
	lastIdxErr indexError
}

func (c *Controller) SetProcedures(apiver uint8, module moduletype, fn func(proc procedure, id uint8, arg []byte) error) {
	c.procedures[uint32(apiver)<<16|uint32(module)] = fn
}

func (c *Controller) CallProcedure(apiver uint8, module moduletype, proc procedure, id uint8, arg []byte) error {
	p, ok := c.procedures[uint32(apiver)<<16|uint32(module)]
	if !ok || p == nil {
		return errors.New("no matching procedure found")
	}
	return p(proc, id, arg)
}

func (c *Controller) VisitProcedures(proc *Process, fn func(apiver uint8, module moduletype, proc procedure, id uint8, args []byte) error) error {
	if proc.ConfigCRC != c.cfg.CRC {
		return errors.New("process config does not match controller config")
	}
	err := proc.VisitProcessUnits(0, func(offsetIdx uint16, u Unit) error {
		if int(u.Sequence) >= len(c.cfg.Sequences) {
			return c.makeIdxErr("sequence index", int(u.Sequence), len(c.cfg.Sequences))
		}
		seq := &c.cfg.Sequences[u.Sequence]
		for _, scmd := range seq.Commands {
			cl := c.cfg.GetCommandList(scmd.APIVersion, scmd.ModuleType)
			if cl == nil {
				return errors.New("no command list found for moduletype and API version")
			} else if int(scmd.CommandIndex) >= len(cl.Commands) {
				return c.makeIdxErr("command index", int(scmd.CommandIndex), len(cl.Commands))
			}
			cmd := cl.Commands[scmd.CommandIndex]
			err := fn(scmd.APIVersion, scmd.ModuleType, cmd.Procedure, scmd.ModuleID, cmd.Args)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) Exec(proc *Process) error {
	return c.VisitProcedures(proc, c.CallProcedure)
}

func (c *Controller) makeIdxErr(msg string, idx, lim int) *indexError {
	c.lastIdxErr = indexError{
		Msg:   msg,
		Index: idx,
		Limit: lim,
	}
	return &c.lastIdxErr
}

type indexError struct {
	Msg   string
	Index int
	Limit int
}

func (e indexError) Error() string {
	return fmt.Sprintf("%s: %d out of range (limit: %d)", e.Msg, e.Index, e.Limit)
}
