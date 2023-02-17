package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Server struct {
	IP   string
	Port string
}

type Path struct {
	FfmpegPath       string `mapstructure:"ffmpeg_path"`
	StaticSourcePath string `mapstructure:"static_source_path"`
}

type MySQL struct {
	Host      string
	Port      string
	Database  string
	Username  string
	Password  string
	Charset   string
	ParseTime bool `mapstructure:"parse_time"`
	Loc       string
}

type Redis struct {
	Host     string
	Port     string
	Database int
}

type COS struct {
	BucketName string `mapstructure:"bucket_name"`
	AppID      string `mapstructure:"app_id"`
	Region     string
	SecretID   string `mapstructure:"secret_id"`
	SecretKey  string `mapstructure:"secret_key"`
	URLFormat  string `mapstructure:"url_format"`
}

type Config struct {
	DB     MySQL `mapstructure:"mysql"`
	RDB    Redis `mapstructure:"redis"`
	Server `mapstructure:"server"`
	Path   `mapstructure:"path"`
	COS    `mapstructure:"cos"`
}

var Info Config

func InitEnv() {
	viper.SetConfigFile("C:/Users/sysu/simple-tiktok/config/config.toml") // 指定配置文件路径
	err := viper.ReadInConfig()                                           // 读取配置信息
	if err != nil {                                                       // 读取配置信息失败
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	if err := viper.Unmarshal(&Info); err != nil {
		panic(fmt.Errorf("unmarshal conf failed, err:%s \n", err))
	}
	fmt.Printf("Server: %+v\n", Info.Server)
	fmt.Printf("path: %+v\n", Info.Path)
	fmt.Printf("MySQL: %+v\n", Info.DB)
	fmt.Printf("Redis: %+v\n", Info.RDB)
	fmt.Printf("COS: %+v\n", Info.COS)
}
