package config

const (
	Dev = "dev"
	Pre = "pre"
	Prd = "prd"
)

type AppConfig struct {
	HttpPort uint64    `yaml:"http_port"`
	Env      string    `yaml:"env"`
	EthUrl   string    `yaml:"eth_url"`
	Secret   string    `yaml:"secret"`
	Owner    string    `yaml:"owner"`
	Token    *Token    `yaml:"token"`
	DB       *Dyanmodb `yaml:"db"`
	SQS      *SQS      `yaml:"sqs"`
}
type Token struct {
	FilePath string `yaml:"file_path"`
	Address  string `yaml:"address"`
	Symbol   string `yaml:"symbol"`
	Decimals int    `yaml:"decimals"`
}
type Dyanmodb struct {
	Host   string `yaml:"host"`
	Port   uint64 `yaml:"port"`
	Region string `yaml:"region"`
	Table  string `yaml:"table"`
}

type SQS struct {
	Host   string `yaml:"host"`
	Port   uint64 `yaml:"port"`
	Region string `yaml:"region"`
	URL    string `yaml:"url"`
}

func (cfg *AppConfig) IsDevEnv() bool {
	return cfg.Env == "dev"
}
