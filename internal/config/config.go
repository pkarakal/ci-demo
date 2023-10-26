package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DefaultConfigDirs = []string{
	".",
	"/",
	"./config",
	"/etc/ci-demo",
}

type Configuration struct {
	Settings ConfigurationSettings `json:"settings"`
}

type ConfigurationSettings struct {
	Port           uint           `json:"port"`
	Adapter        string         `json:"adapter"`
	AdapterOptions AdapterOptions `json:"adapterOptions"`
}

type AdapterOptions struct {
	Server   string             `json:"server"`
	Port     int64              `json:"port"`
	Database string             `json:"database"`
	UseTLS   bool               `json:"useTLS"`
	Auth     AdapterAuthOptions `json:"auth"`
}

type AdapterAuthOptions struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func SetDefaults() {
	viper.SetDefault(ListenPort, 5000)
	viper.SetDefault(DatabaseUseTLS, true)
	viper.SetDefault(DatabaseHost, "localhost")
	viper.SetDefault(DatabasePort, 5432)
	viper.SetDefault(DatabaseType, "postgres")
	viper.SetDefault(DatabaseUsername, "demo")
	viper.SetDefault(DatabaseName, "demo")
}

func (c *Configuration) OpenDatabase() (*gorm.DB, error) {
	switch c.Settings.Adapter {
	case PostgresDatabase:
		return gorm.Open(postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Athens",
			c.Settings.AdapterOptions.Server,
			c.Settings.AdapterOptions.Auth.Username,
			c.Settings.AdapterOptions.Auth.Password,
			c.Settings.AdapterOptions.Database,
			c.Settings.AdapterOptions.Port)))
	case MySQLDatabase:
		return gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Settings.AdapterOptions.Auth.Username,
			c.Settings.AdapterOptions.Auth.Password,
			c.Settings.AdapterOptions.Server,
			c.Settings.AdapterOptions.Port,
			c.Settings.AdapterOptions.Database)))
	default:
		return nil, errors.New(ErrAdapterType)
	}
}

func SetUpEnvVarReader() {
	// read from environment variables
	viper.AutomaticEnv()

	// Replace "." with "_" when reading environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func LoadConfig() (*Configuration, error) {
	SetDefaults()
	SetUpEnvVarReader()

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	for _, dir := range DefaultConfigDirs {
		viper.AddConfigPath(dir)
	}
	err := viper.ReadInConfig()
	// It's ok if the config file doesn't exist, but we want to catch any
	// other config-related issues
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file %v", err)
		}

		zap.L().Info("no config file found, proceeding without one")
	}

	config := &Configuration{}

	err = viper.Unmarshal(config)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Couldn't unmarshal config into struct, %v", err), zap.Error(err))
		return nil, err
	}
	return config, nil
}

func InitLogging() (*zap.Logger, func()) {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.WarnLevel)
	// initialize logger
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(
			zap.NewProductionEncoderConfig()),
		zapcore.Lock(os.Stdout), atomicLevel,
	),
	)

	undo := zap.ReplaceGlobals(logger)
	return logger, undo
}
