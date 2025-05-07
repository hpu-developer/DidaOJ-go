package service

import (
	"meta/singleton"
)

type MigrateContestService struct{}

var singletonMigrateContestService = singleton.Singleton[MigrateContestService]{}

func GetMigrateContestService() *MigrateContestService {
	return singletonMigrateContestService.GetInstance(
		func() *MigrateContestService {
			return &MigrateContestService{}
		},
	)
}

func (s *MigrateContestService) Start() error {

	//ctx := context.Background()
	//jolMysqlClient := metamysql.GetSubsystem().GetClient("jol")

	return nil
}
