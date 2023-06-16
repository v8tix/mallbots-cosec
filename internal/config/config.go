package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type (
	PGConfig struct {
		Conn string `json:"uri,omitempty"`
	}

	NatsConfig struct {
		URL        string `json:"url,omitempty"`
		Stream     string `json:"stream,omitempty"`
		Username   string `json:"username,omitempty"`
		Password   string `json:"password,omitempty"`
		ClientName string `json:"client_name,omitempty"`
	}

	AppConfig struct {
		Environment     string        `json:"environment,omitempty"`
		LogLevel        string        `json:"log_level,omitempty"`
		PG              PGConfig      `json:"db_cfg,omitempty"`
		Nats            NatsConfig    `json:"nats_cfg,omitempty"`
		RPC             RPCConfig     `json:"rpc_cfg,omitempty"`
		Web             WebConfig     `json:"web_cfg,omitempty"`
		ShutdownTimeout time.Duration `json:"shutdown_timeout,omitempty"`
	}
)

type RPCConfig struct {
	Host string `json:"host,omitempty"`
	Port string `json:"port,omitempty"`
}

func (c RPCConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type WebConfig struct {
	Host string `json:"host,omitempty"`
	Port string `json:"port,omitempty"`
}

func (c WebConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func InitConfig(configFile string, cfg *AppConfig) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	return nil
}
