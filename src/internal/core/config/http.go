package config

const (
	portEnvName = "HTTP_PORT"
	hostEnvName = "HTTP_HOST"
)

type HTTPConfig interface {
	Port() string
	Host() string
}

type httpConfig struct {
	port string
	host string
}

func NewHTTPConfig() (HTTPConfig, error) {
	port := getEnv(portEnvName, "8080")
	host := getEnv(hostEnvName, "localhost")
	return &httpConfig{
		port: port,
		host: host,
	}, nil
}

func (cfg *httpConfig) Port() string {
	return cfg.port
}

func (cfg *httpConfig) Host() string {
	return cfg.host
}
