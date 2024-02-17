package pcc

func (cfg *ControllerConfig) AddCommand(apiver uint8, module moduletype, cmd Command) {
	cl := cfg.GetCommandList(apiver, module)
	if cl != nil {
		cl.Commands = append(cl.Commands, cmd)
		return
	}
	// No existing command list for this module, create a new one.
	cfg.CommandLists = append(cfg.CommandLists, CommandList{
		APIVersion: apiver,
		ModuleType: module,
		Commands:   []Command{cmd},
	})
}

// GetCommandList returns the command list for the given API version and module type.
func (cfg ControllerConfig) GetCommandList(apiver uint8, module moduletype) *CommandList {
	for i := range cfg.CommandLists {
		cl := &cfg.CommandLists[i]
		if cl.APIVersion == apiver && cl.ModuleType == module {
			return cl
		}
	}
	return nil
}
