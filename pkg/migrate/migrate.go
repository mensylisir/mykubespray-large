package migrate

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/utils"
)

const (
	phaseName = "migrate"
)

const (
	DefaultPassword = "Def@u1tpwd"
)

const (
	releaseMigrationDir = "/usr/local/lib/middleware/migration"
	localMigrationDir   = "./migration"
)

var migrationDirs = []string{
	localMigrationDir,
	releaseMigrationDir,
}

type InitMigrateDBPhase struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

func (i *InitMigrateDBPhase) Init() error {
	aesPasswd, er1 := utils.StringEncrypt(i.Password)
	if er1 != nil {
		return er1
	}
	p, err := utils.StringDecrypt(aesPasswd)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=Asia%%2FShanghai",
		i.User,
		p,
		i.Host,
		i.Port,
		i.Name)
	var path string
	for _, d := range migrationDirs {
		if utils.Exists(d) {
			path = d
		}
	}
	if path == "" {
		return fmt.Errorf("can not find migration in [%s,%s]", localMigrationDir, releaseMigrationDir)
	}
	filePath := fmt.Sprintf("file://%s", path)
	m, err := migrate.New(
		filePath, url)
	if err != nil {
		logger.GetLogger().Errorf("Failed to migrate database info[%s,%s]: %s", filePath, url, err.Error())
		return err
	}
	// 初始化默认用户
	_, _, _ = m.Version()
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.GetLogger().Info("no databases change,skip migrate")
			return nil
		}
		return err
	}
	_, err = utils.StringEncrypt(DefaultPassword)
	if err != nil {
		logger.GetLogger().Errorf("Failed to init default user: %s", err.Error())
		return fmt.Errorf("can not init default user")
	}
	//if !(v > 0) {
	//	if err := db.DB.Model(&model.User{}).Where("name = ?", "admin").Updates(map[string]interface{}{"Password": dp}).Error; err != nil {
	//		logger.Log.Errorf("Failed to  update default use: %s", err.Error())
	//		return fmt.Errorf("can not update default user")
	//	}
	//}
	return nil
}

func (i *InitMigrateDBPhase) PhaseName() string {
	return phaseName
}
