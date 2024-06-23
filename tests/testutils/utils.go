package testutils

import (
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/container"
)

func GetConfig(configDir string) *config.Config {
	cfg, err := config.LoadYaml(configDir)
	if err != nil {
		panic(err)
	}

	return cfg
}
func GetContainer(configDir string) *container.Container {
	return container.New(GetConfig(configDir))
}
