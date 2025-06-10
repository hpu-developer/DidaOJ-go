package application

import (
	"meta/engine"
	"meta/subsystem"
)

type Subsystem struct {
	subsystem.Subsystem
}

func GetSubsystem() *Subsystem {
	if thisSubsystem := engine.GetSubsystem[*Subsystem](); thisSubsystem != nil {
		return thisSubsystem.(*Subsystem)
	}
	return nil
}

func (s *Subsystem) GetName() string {
	return "Migrate"
}

func (s *Subsystem) Start() error {
	err := s.startSubSystem()
	if err != nil {
		return err
	}
	return nil
}

func (s *Subsystem) startSubSystem() error {
	return nil
}
