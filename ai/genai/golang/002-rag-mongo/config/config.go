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

	// MongoDB Configuration
	MongoURI          string `mapstructure:"mongo_uri"`
	MongoDatabase     string `mapstructure:"mongo_database"`
	MongoCollection   string `mapstructure:"mongo_collection"`
	MongoVectorField  string `mapstructure:"mongo_vector_field"`
	MongoContentField string `mapstructure:"mongo_content_field"`

	// Gemini API Key
	GeminiAPIKey string `mapstructure:"gemini_api_key"`
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

	viper.AddConfigPath(fc.ConfigPath)
	viper.SetConfigName(fc.FileName)
	viper.SetConfigType(fc.Extentsion)
	viper.AutomaticEnv()

	// Bind environment variables
	viper.BindEnv("mongo_uri", "MONGO_URI")
	viper.BindEnv("mongo_database", "MONGO_DATABASE")
	viper.BindEnv("mongo_collection", "MONGO_COLLECTION")
	viper.BindEnv("mongo_vector_field", "MONGO_VECTOR_FIELD")
	viper.BindEnv("mongo_content_field", "MONGO_CONTENT_FIELD")
	viper.BindEnv("gemini_api_key", "GEMINI_API_KEY")

	// Set defaults
	viper.SetDefault("mongo_uri", "")
	viper.SetDefault("mongo_database", "")
	viper.SetDefault("mongo_collection", "")
	viper.SetDefault("mongo_vector_field", "")
	viper.SetDefault("mongo_content_field", "")
	viper.SetDefault("gemini_api_key", "")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err.(viper.ConfigFileNotFoundError)
		}
		return nil, err
	}

	var appConfigs Configs
	err = viper.Unmarshal(&appConfigs)
	if err != nil {
		return nil, err
	}

	// Populate the Fields map with all settings from Viper for flexibility
	var allSettings map[string]interface{}
	if err := viper.Unmarshal(&allSettings); err != nil {
		return nil, err
	}
	appConfigs.Fields = allSettings

	// AppConfig.ConfigFile = AppConfig.ConfigFilePath + "config.yaml"
	err = os.Setenv("JSON_CONFIG_PATH", fc.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	appConfigs.PathConfigFile = fc.ConfigFilePath
	appConfigs.Paths = &AppConfig
	appConfigs.FileConfig = &fc

	return &appConfigs, err
}
