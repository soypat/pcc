package pcc

import "errors"

func (p *Process) VisitProcessUnits(startIdx uint16, visitor func(idx uint16, u Unit) error) error {
	if len(p.Units) == 0 {
		return errors.New("no units to visit")
	} else if len(p.Units) > 63 {
		return errors.New("too many units to visit") // overflows bitmask.
	} else if startIdx >= uint16(len(p.Units)) {
		return errors.New("start index out of range")
	}
	var nxtIdx uint16 = startIdx
	// bitmask to track visited units.
	var visited uint64
	for {
		currentUnit := p.Units[nxtIdx]
		endIter := currentUnit.Flags&unitFlagEndMask != 0
		if !endIter && currentUnit.Next == 0 {
			return errors.New("no next unit (zero idx)")
		}
		err := visitor(nxtIdx, currentUnit)
		if err != nil {
			return err
		} else if endIter {
			return nil // Normal end to visitation.
		}
		nxtIdx = currentUnit.Next
		if (1<<nxtIdx)&visited != 0 {
			return errors.New("circular reference detected")
		} else if int(nxtIdx) >= len(p.Units) {
			return errors.New("next unit index out of range")
		}
		visited |= 1 << nxtIdx
		currentUnit = p.Units[nxtIdx]
	}
}
