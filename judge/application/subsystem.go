package application

import (
	"judge/service"
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
	return "Judge"
}

func (s *Subsystem) Start() error {
	err := s.startSubSystem()
	if err != nil {
		return err
	}
	return nil
}

func (s *Subsystem) startSubSystem() error {

	var err error

	err = service.GetStatusService().Start()
	if err != nil {
		return err
	}

	err = service.GetJudgeService().Start()
	if err != nil {
		return err
	}

	err = service.GetRemoteService().Start()
	if err != nil {
		return err
	}

	return nil
}
