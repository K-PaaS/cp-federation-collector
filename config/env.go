package config

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"
)

var Env *envConfigs

func init() {
	Env = loadEnvVariables()
}

type envConfigs struct {
	HostClusterName string `mapstructure:"HostClusterName"`
	KarmadaApi      string `mapstructure:"KarmadaApi"`
	KarmadaToken    string `mapstructure:"KarmadaToken"`
	NatsBucketName  string `mapstructure:"NatsBucketName"`
	NatsId          string `mapstructure:"NatsId"`
	NatsPassword    string `mapstructure:"NatsPassword"`
	NatsSubjectName string `mapstructure:"NatsSubjectName"`
	NatsUrl         string `mapstructure:"NatsUrl"`
	VaultRoleId     string `mapstructure:"VaultRoleId"`
	VaultSecretId   string `mapstructure:"VaultSecretId"`
	VaultUrl        string `mapstructure:"VaultUrl"`
}

func loadEnvVariables() (config *envConfigs) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading env file", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}
	return
}
