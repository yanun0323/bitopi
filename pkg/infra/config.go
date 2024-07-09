package infra

import (
	"bitopi/internal/util"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/spf13/viper"
	"github.com/yanun0323/pkg/config"
)

var _once sync.Once

func Init(cfgName string) error {
	var err error
	_once.Do(
		func() {
			_, f, _, _ := runtime.Caller(0)
			cfgPath := filepath.Join(filepath.Dir(f), "../../config")
			if err = config.Init(cfgPath, cfgName, true); err != nil {
				return
			}
		},
	)
	if err != nil {
		return err
	}
	viper.AddConfigPath(util.GetAbsPath("config"))
	viper.AddConfigPath(util.GetAbsPath())
	viper.AutomaticEnv()
	viper.SetConfigName(cfgName)
	viper.SetConfigType("yaml")

	return viper.ReadInConfig()
}
