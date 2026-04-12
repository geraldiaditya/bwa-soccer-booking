package utils

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func BindFromJSON(dest any, filename, path string) error {
	v := viper.New()
	v.SetConfigType("json")
	v.AddConfigPath(path)
	v.SetConfigName(filename)

	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}
	return nil
}

