package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

var AppConfig ConfigPath

type Timeout struct{}

type FileConfig struct {
	Extentsion     string
	FileName       string
	ConfigPath     string
	ConfigFilePath string
}

type ConfigPath struct {
	RootPath       string
	CmdPath        string
	ConfigsPath    string
	PkgPath        string
	InternalPath   string
	ServicePath    string
	DomainPath     string
	RepositoryPath string
	AdapterPath    string
	HandlerPath    string
	SwaggerPath    string
	InfraPath      string
}

func init() {
	_, filename, _, _ := runtime.Caller(0)

	// Root directory
	AppConfig.RootPath = filepath.Dir(filepath.Dir(filename))

	// System paths
	AppConfig.ConfigsPath = (filepath.Join(AppConfig.RootPath, "configs"))
	AppConfig.CmdPath = (filepath.Join(AppConfig.RootPath, "cmd"))
	AppConfig.InternalPath = (filepath.Join(AppConfig.RootPath, "internal"))
	AppConfig.PkgPath = (filepath.Join(AppConfig.RootPath, "pkg"))
	AppConfig.InfraPath = (filepath.Join(AppConfig.RootPath, "infra"))
	AppConfig.SwaggerPath = (filepath.Join(AppConfig.RootPath, "docs"))

	// System paths inside internal
	AppConfig.ServicePath = (filepath.Join(AppConfig.InternalPath, "service"))
	AppConfig.DomainPath = (filepath.Join(AppConfig.InternalPath, "entity"))
	AppConfig.RepositoryPath = (filepath.Join(AppConfig.InternalPath, "repository"))
	AppConfig.AdapterPath = (filepath.Join(AppConfig.InternalPath, "adapter"))
	AppConfig.HandlerPath = (filepath.Join(AppConfig.InternalPath, "handler"))
}

type Configs struct {
	// Provider Provider `json:"provider" binding:"required" mapstructure:"provider"`
	PathConfigFile string `mapstructure:"path_config_file"`
	Fields         map[string]interface{}
	Paths          *ConfigPath
	FileConfig     *FileConfig
}

func LoadConfig() (*Configs, error) {

	path := os.Getenv("PATH_CONFIG")
	if path == "" {
		log.Println("variável PATH_CONFIG para informar o path do arquivo config.yaml não foi informado")
		path = AppConfig.CmdPath + "/.config"
	}
	fc := FileConfig{
		ConfigPath:     path,
		Extentsion:     "yaml",
		FileName:       "config",
		ConfigFilePath: path + "/config" + ".yaml",
	}

	var cfg interface{}

	viper.AddConfigPath(fc.ConfigPath)
	viper.SetConfigName(fc.FileName)
	viper.SetConfigType(fc.Extentsion)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err.(viper.ConfigFileNotFoundError)
		}
		return nil, err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	// AppConfig.ConfigFile = AppConfig.ConfigFilePath + "config.yaml"
	err = os.Setenv("JSON_CONFIG_PATH", fc.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	return &Configs{
		PathConfigFile: fc.ConfigFilePath,
		Paths:          &AppConfig,
		FileConfig:     &fc,
		Fields:         cfg.(map[string]interface{}),
	}, err
}
