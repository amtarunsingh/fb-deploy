package config

import (
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/timeutil"
	env "github.com/caarlos0/env/v10"
)

const (
	CountersTtlHours                    = 48
	ProjectName                         = "User Votes Storage"
	ProjectVersion                      = "1.0.0"
	DynamoDbVersionConflictRetriesCount = 3
)

type RomancesConfig struct {
	MutualRomanceTtlSeconds    int64
	NonMutualRomanceTtlSeconds int64
	DeadRomanceTtlSeconds      int64
}

type CountersConfig struct {
	TtlSeconds int64
}

type Config struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"INFO"`
	Aws      struct {
		Region                string `env:"AWS_REGION" envDefault:"us-east-2"`
		AccessKeyId           string `env:"AWS_ACCESS_KEY_ID" envDefault:"dummy"`
		SecretAccessKey       string `env:"AWS_SECRET_ACCESS_KEY" envDefault:"dummy"`
		DynamoDbLocalEndpoint string `env:"DYNAMO_DB_ENDPOINT"`
		SnsLocalEndpoint      string `env:"SNS_DB_ENDPOINT"`
	}
	Counters CountersConfig
	Romances RomancesConfig
}

type ServerOptions struct {
	Host string `doc:"Hostname to listen on." default:"0.0.0.0"`
	Port int    `doc:"Port to listen on." short:"p" default:"8888"`
}

func Load() Config {
	cfg := Config{
		Counters: CountersConfig{
			TtlSeconds: CountersTtlHours * timeutil.HourSeconds,
		},
		Romances: RomancesConfig{
			MutualRomanceTtlSeconds:    546 * timeutil.DaySeconds,
			NonMutualRomanceTtlSeconds: 180 * timeutil.DaySeconds,
			DeadRomanceTtlSeconds:      90 * timeutil.DaySeconds,
		},
	}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return cfg
}
