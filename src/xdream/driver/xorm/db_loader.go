package xorm

import (
	"errors"
)

//加载数据库配置
func InitWithConfig(config map[string]map[string]DbConfig) error {
	return RegisterDbsWithConfig(config)
}

func RegisterDbsWithConfig(dbConfigs map[string]map[string]DbConfig) error {

	for dbName, dbConfig := range dbConfigs {
		masterConfig, ok := dbConfig["master"]
		if !ok {
			return errors.New("Db config must have a master")
		}

		if err := RegisterDb(dbName, &masterConfig); err != nil {
			return err
		}

		if slaveConfig, ok := dbConfig["slave"]; ok {
			if err := RegisterSlaveDb(dbName, &slaveConfig); err != nil {
				return err
			}
		}
	}

	return nil
}
