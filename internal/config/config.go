package config

import (
	"strings"
	"sync"

	"github.com/spf13/pflag"

	"github.com/superfly/flyctl/internal/env"
	"github.com/superfly/flyctl/internal/flag/flagnames"
)

const (
	// FileName denotes the name of the config file.
	FileName = "config.yml"

	envKeyPrefix          = "FLY_"
	apiBaseURLEnvKey      = envKeyPrefix + "API_BASE_URL"
	flapsBaseURLEnvKey    = envKeyPrefix + "FLAPS_BASE_URL"
	metricsBaseURLEnvKey  = envKeyPrefix + "METRICS_BASE_URL"
	AccessTokenEnvKey     = envKeyPrefix + "ACCESS_TOKEN"
	AccessTokenFileKey    = "access_token"
	MetricsTokenEnvKey    = envKeyPrefix + "METRICS_TOKEN"
	MetricsTokenFileKey   = "metrics_token"
	SendMetricsFileKey    = "send_metrics"
	WireGuardStateFileKey = "wire_guard_state"
	APITokenEnvKey        = envKeyPrefix + "API_TOKEN"
	orgEnvKey             = envKeyPrefix + "ORG"
	registryHostEnvKey    = envKeyPrefix + "REGISTRY_HOST"
	organizationEnvKey    = envKeyPrefix + "ORGANIZATION"
	regionEnvKey          = envKeyPrefix + "REGION"
	verboseOutputEnvKey   = envKeyPrefix + "VERBOSE"
	jsonOutputEnvKey      = envKeyPrefix + "JSON"
	logGQLEnvKey          = envKeyPrefix + "LOG_GQL_ERRORS"
	localOnlyEnvKey       = envKeyPrefix + "LOCAL_ONLY"

	defaultAPIBaseURL     = "https://api.fly.io"
	defaultFlapsBaseURL   = "https://api.machines.dev"
	defaultRegistryHost   = "registry.fly.io"
	defaultMetricsBaseURL = "https://flyctl-metrics.fly.dev"
)

// Config wraps the functionality of the configuration file.
//
// Instances of Config are safe for concurrent use.
type Config struct {
	mu sync.RWMutex

	// APIBaseURL denotes the base URL of the API.
	APIBaseURL string

	// FlapsBaseURL denotes base URL for FLAPS (also known as the Machines API).
	FlapsBaseURL string

	// MetricsBaseURL denotes the base URL of the metrics API.
	MetricsBaseURL string

	// RegistryHost denotes the docker registry host.
	RegistryHost string

	// VerboseOutput denotes whether the user wants the output to be verbose.
	VerboseOutput bool

	// JSONOutput denotes whether the user wants the output to be JSON.
	JSONOutput bool

	// LogGQLErrors denotes whether the user wants the log GraphQL errors.
	LogGQLErrors bool

	// SendMetrics denotes whether the user wants to send metrics.
	SendMetrics bool

	// Organization denotes the organizational slug the user has selected.
	Organization string

	// Region denotes the region slug the user has selected.
	Region string

	// LocalOnly denotes whether the user wants only local operations.
	LocalOnly bool

	// AccessToken denotes the user's access token.
	AccessToken string

	// MetricsToken denotes the user's metrics token.
	MetricsToken string
}

// New returns a new instance of Config populated with default values.
func New() *Config {
	return &Config{
		APIBaseURL:     defaultAPIBaseURL,
		FlapsBaseURL:   defaultFlapsBaseURL,
		RegistryHost:   defaultRegistryHost,
		MetricsBaseURL: defaultMetricsBaseURL,
	}
}

// ApplyEnv sets the properties of cfg which may be set via environment
// variables to the values these variables contain.
//
// ApplyEnv does not change the dirty state of config.
func (cfg *Config) ApplyEnv() {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	cfg.AccessToken = env.FirstOrDefault(cfg.AccessToken,
		AccessTokenEnvKey, APITokenEnvKey)

	// trim whitespace since it causes http errors when passsed to Docker auth
	cfg.AccessToken = strings.TrimSpace(cfg.AccessToken)

	cfg.VerboseOutput = env.IsTruthy(verboseOutputEnvKey) || cfg.VerboseOutput
	cfg.JSONOutput = env.IsTruthy(jsonOutputEnvKey) || cfg.JSONOutput
	cfg.LogGQLErrors = env.IsTruthy(logGQLEnvKey) || cfg.LogGQLErrors
	cfg.LocalOnly = env.IsTruthy(localOnlyEnvKey) || cfg.LocalOnly

	cfg.Organization = env.FirstOrDefault(cfg.Organization,
		orgEnvKey, organizationEnvKey)
	cfg.Region = env.FirstOrDefault(cfg.Region, regionEnvKey)
	cfg.RegistryHost = env.FirstOrDefault(cfg.RegistryHost, registryHostEnvKey)
	cfg.APIBaseURL = env.FirstOrDefault(cfg.APIBaseURL, apiBaseURLEnvKey)
	cfg.FlapsBaseURL = env.FirstOrDefault(cfg.FlapsBaseURL, flapsBaseURLEnvKey)
	cfg.MetricsBaseURL = env.FirstOrDefault(cfg.MetricsBaseURL, metricsBaseURLEnvKey)
}

// ApplyFile sets the properties of cfg which may be set via configuration file
// to the values the file at the given path contains.
func (cfg *Config) ApplyFile(path string) (err error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	var w struct {
		AccessToken  string `yaml:"access_token"`
		MetricsToken string `yaml:"metrics_token"`
		SendMetrics  bool   `yaml:"send_metrics"`
	}
	w.SendMetrics = true

	if err = unmarshal(path, &w); err == nil {
		cfg.AccessToken = w.AccessToken
		cfg.MetricsToken = w.MetricsToken
		cfg.SendMetrics = w.SendMetrics
	}

	return
}

// ApplyFlags sets the properties of cfg which may be set via command line flags
// to the values the flags of the given FlagSet may contain.
func (cfg *Config) ApplyFlags(fs *pflag.FlagSet) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	applyStringFlags(fs, map[string]*string{
		flagnames.AccessToken: &cfg.AccessToken,
		flagnames.Org:         &cfg.Organization,
		flagnames.Region:      &cfg.Region,
	})

	applyBoolFlags(fs, map[string]*bool{
		flagnames.Verbose:    &cfg.VerboseOutput,
		flagnames.JSONOutput: &cfg.JSONOutput,
		flagnames.LocalOnly:  &cfg.LocalOnly,
	})
}

func (cfg *Config) MetricsBaseURLIsProduction() bool {
	return cfg.MetricsBaseURL == defaultMetricsBaseURL
}

func applyStringFlags(fs *pflag.FlagSet, flags map[string]*string) {
	for name, dst := range flags {
		if !fs.Changed(name) {
			continue
		}

		if v, err := fs.GetString(name); err != nil {
			panic(err)
		} else {
			*dst = v
		}
	}
}

func applyBoolFlags(fs *pflag.FlagSet, flags map[string]*bool) {
	for name, dst := range flags {
		if !fs.Changed(name) {
			continue
		}

		if v, err := fs.GetBool(name); err != nil {
			panic(err)
		} else {
			*dst = v
		}
	}
}
