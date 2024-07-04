package testutils

import (
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/container"
	log "github.com/sirupsen/logrus"
)

func GetConfig(configDir string) *config.Config {
	cfg, err := config.LoadYaml(configDir)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
func GetContainer(configDir string) *container.Container {
	return container.New(GetConfig(configDir))
}
