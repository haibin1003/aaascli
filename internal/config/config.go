package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 配置结构
type Config struct {
	Cookie string `json:"cookie"`
}

// GetDefaultConfigPath 返回默认配置文件路径
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(homeDir, ".sdp", "config.json")
}

// EnsureConfigDir 确保配置目录存在
func EnsureConfigDir() error {
	configPath := GetDefaultConfigPath()
	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{}
}

// LoadConfigWithDefaults 加载配置
func LoadConfigWithDefaults(filename string) (*Config, error) {
	cfg := NewConfig()

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		// 尝试解析简单格式 {"cookie": "xxx"}
		var simple map[string]string
		if err := json.Unmarshal(data, &simple); err == nil {
			if cookie, ok := simple["cookie"]; ok {
				cfg.Cookie = cookie
			}
		}
	}

	return cfg, nil
}

// SaveConfig 保存配置
func SaveConfig(cfg *Config) error {
	configPath := GetDefaultConfigPath()

	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("ensure config dir failed: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config failed: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file failed: %w", err)
	}

	return nil
}
