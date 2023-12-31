package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"

	"github.com/r-mol/balanser_highload_system/internal/balancer"
	"github.com/r-mol/balanser_highload_system/internal/proxy"
)

type BalancerConfig struct {
	Servers []ServerConfig `yaml:"servers"`
}

type ServerConfig struct {
	URL      string `yaml:"url"`
	Priority int32  `yaml:"priority"`
}

func validateBalancerConfig(config BalancerConfig) error {
	switch {
	case config.Servers == nil:
		return fmt.Errorf("\"servers\" is not provided")
	}

	for _, server := range config.Servers {
		if err := validateServerConfig(server); err != nil {
			return fmt.Errorf("\"server\": %w", err)
		}
	}

	return nil
}

func validateServerConfig(config ServerConfig) error {
	switch {
	case config.URL == "":
		return fmt.Errorf("\"url\" is not provided")
	case config.Priority == 0:
		return fmt.Errorf("\"priority\" is not provided")
	}

	return nil
}

func ParseBalancerConfig(path string) (BalancerConfig, error) {
	config := BalancerConfig{}
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to load %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("unable to unmarshal: %w", err)
	}

	if err := validateBalancerConfig(config); err != nil {
		return config, fmt.Errorf("invalid data: %w", err)
	}

	return config, nil
}

func GetBalancerFromConfig(config BalancerConfig, logger *log.Logger, metrics *balancer.Metrics) (*balancer.LoadBalancer, error) {
	proxies := map[*proxy.Proxy]int32{}
	for _, server := range config.Servers {
		u, err := url.Parse(server.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse url: %w", err)
		}

		p, err := proxy.New(proxy.WithOrigin(u))
		if err != nil {
			return nil, fmt.Errorf("failed to create new halth proxy: %w", err)
		}

		proxies[p] = server.Priority
	}

	b, err := balancer.New(balancer.WithProxies(proxies), balancer.WithLogger(logger), balancer.WithMetrics(metrics))
	if err != nil {
		return nil, fmt.Errorf("failed to get new balancer: %w", err)
	}

	return b, nil
}
