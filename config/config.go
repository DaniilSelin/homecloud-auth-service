package config

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// LoggerConfig - конфигурация логгера
type LoggerConfig struct {
	zap.Config `yaml:",inline"`
}

func (lc *LoggerConfig) Build() (*zap.Logger, error) {
	return lc.Config.Build()
}

// ServerConfig - конфигурация HTTP сервера
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// JwtConfig - конфигурация JWT токенов
type JwtConfig struct {
	SecretKey  string        `yaml:"secret_key"`
	Expiration time.Duration `yaml:"expiration"`
}

// VerificationConfig - конфигурация токенов верификации
type VerificationConfig struct {
	SecretKey  string        `yaml:"secret_key"`
	Expiration time.Duration `yaml:"expiration"`
}

// GrpcConfig - конфигурация gRPC клиента для БД
type GrpcConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DbManagerConfig - конфигурация gRPC клиента для DBManager
type DbManagerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Config - основная конфигурация приложения
type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Jwt          JwtConfig          `yaml:"jwt"`
	Verification VerificationConfig `yaml:"verification"`
	Logger       LoggerConfig       `yaml:"logger"`
	Grpc         GrpcConfig         `yaml:"grpc"`
	DbManager    DbManagerConfig    `yaml:"dbmanager"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("could not decode config file: %v", err)
	}
	return &config, nil
}
