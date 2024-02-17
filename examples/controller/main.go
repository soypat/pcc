package main

import (
	"fmt"

	"github.com/soypat/pcc"
)

const (
	modulePrinter = 0xc4fe

	cmdHello = 0
	cmdBye   = 1
)

const (
	procPrint = iota
	procPrintNewline
)

var config = pcc.ControllerConfig{
	Sequences: []pcc.Sequence{
		pcc.Sequence{
			Name: [8]byte{'h', 'e', 'l', 'l', 'o', 'b', 'y', 'e'},
			Commands: []pcc.SequenceCommand{
				{CommandIndex: cmdHello, ModuleType: modulePrinter},
				{CommandIndex: cmdBye, ModuleType: modulePrinter},
			},
		},
	},
	CommandLists: []pcc.CommandList{
		pcc.CommandList{ModuleType: modulePrinter, Commands: []pcc.Command{
			cmdHello: pcc.Command{Procedure: procPrint, Args: []byte("hello world")},
			cmdBye:   pcc.Command{Procedure: procPrintNewline, Args: []byte("bye world")},
		}},
	},
}

func main() {
	var ctller pcc.Controller
	var printer printerModule
	ctller.SetProcedures(0, modulePrinter, printer.procedure)
}

func (p *printerModule) procedure(proc uint16, id uint8, arg []byte) error {
	switch proc {
	case procPrint:
		p.justPrint(string(arg))
	case procPrintNewline:
		p.printWithNewline(string(arg))
	default:
		return fmt.Errorf("unsupported procedure %d", proc)
	}
	return nil
}

type printerModule struct {
}

func (p *printerModule) justPrint(arg string) {
	fmt.Print(arg)
}

func (p *printerModule) printWithNewline(arg string) {
	fmt.Println(arg)
}
