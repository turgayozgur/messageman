package config

import (
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Config inits from configuration file
type Config struct {
	Mode    string
	Port    string `yaml:"-"`
	Metric  MetricConfig
	Logging struct {
		Level    string
		Humanize bool
	} `yaml:"-"`
	RabbitMQ RabbitMQConfig
	Services []ServiceConfig
}

// Config inits from configuration file
type MetricConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Exporter string `yaml:"exporter"`
}

// RabbitMQConfig inits from configuration file
type RabbitMQConfig struct {
	Url string `yaml:"url"`
}

// ServiceConfig inits from configuration file
type ServiceConfig struct {
	Name      string `yaml:"name"`
	Url       string `yaml:"url"`
	Readiness struct {
		Path string `yaml:"path"`
	}
	Workers []WorkerConfig
}

// WorkerConfig inits from configuration file
type WorkerConfig struct {
	Path  string `yaml:"path"`
	Queue string `yaml:"queue"`
}

const (
	DefaultConfigFile  = "messageman.yml"
	UsageConfigFile    = "a messageman configuration yml path. The default path is the same location of messageman where you run."
	DefaultMode        = "gateway"
	DefaultPort        = "8015"
	DefaultRabbitMQUrl = "amqp://guest:guest@localhost:5672/"
	DefaultLogLevel    = "info"
)

var (
	configFile string
	// Cfg includes main configurations.
	Cfg *Config
)

func init() {
	humanize, err := strconv.ParseBool(getEnv("LOG_HUMANIZE", "true"))
	if err != nil {
		humanize = true
	}
	Cfg = &Config{
		Mode:   DefaultMode,
		Port:   getEnv("MESSAGEMAN_PORT", DefaultPort),
		Metric: MetricConfig{},
		RabbitMQ: RabbitMQConfig{
			Url: DefaultRabbitMQUrl,
		},
		Logging: struct {
			Level    string
			Humanize bool
		}{
			Level:    getEnv("LOG_LEVEL", DefaultLogLevel),
			Humanize: humanize,
		},
	}
}

// Load attempts to parse the given config file and return a Config object.
func Load() error {
	// parse flags.
	flag.StringVar(&configFile, "-config-file", DefaultConfigFile, UsageConfigFile)
	flag.StringVar(&configFile, "c", DefaultConfigFile, UsageConfigFile+" (shorthand)")
	flag.Parse()

	// load configuration from file.
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Warn().Msgf("Can not read the configuration file. %+v", err)
		return err
	}

	log.Debug().Msgf("\n%s", buf)

	err = yaml.Unmarshal(buf, &Cfg)
	if err != nil {
		log.Warn().Msgf("Can not parse the configuration file. %+v", err)
		return err
	}

	return nil
}

func IsSidecar() bool {
	return Cfg.Mode == "sidecar"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
