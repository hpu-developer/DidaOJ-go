package application

import (
	"log/slog"
	"meta/engine"
	"meta/subsystem"
	"migrate/config"
	"migrate/service"
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

	var err error

	err = service.GetMigrateDiscussEojService().Start()
	if err != nil {
		return err
	}

	if true {
		return nil
	}

	err = service.GetMongoInitService().Start()
	if err != nil {
		return err
	}

	if config.GetConfig().OnlyInit {
		return nil
	}

	slog.Info("migrate finished")

	return nil
}
