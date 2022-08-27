package conf

import (
	"os"
	"strings"

	"github.com/botuniverse/go-libonebot"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Token     string                    `json:"token" yaml:"token" mapstructure:"token"`
	EndPoint  string                    `json:"end_point" yaml:"end_point" mapstructure:"end_point"`
	Heartbeat libonebot.ConfigHeartbeat `json:"heartbeat" yaml:"heartbeat" mapstructure:"heartbeat"`
	Comm      libonebot.ConfigComm      `json:"comm" yaml:"comm" mapstructure:"comm"`
}

var (
	config Config
)

func InitConfig(path string) {
	pathDir := strings.TrimSuffix(path, "config.yml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(pathDir)
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Errorln("配置文件不存在" + err.Error())
			data, _ := yaml.Marshal(&config)
			err := os.WriteFile(path, data, 0666)
			if err != nil {
				log.Errorln("写入配置文件错误" + err.Error())
				return
			}
		} else {
			log.Errorln("加载配置文件出现未知错误" + err.Error())
		}
	}
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Errorln("配置文件加载异常" + err.Error())
		return
	}
}

func GetConfig() *Config {
	return &config
}
