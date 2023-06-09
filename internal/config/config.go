package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type BigQueryConfig struct {
	ProjectID       string `yaml:"project_id"`
	CredentialsJSON string `yaml:"credentials_json"`
	Dataset         string `yaml:"dataset"`
	Table           string `yaml:"table"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type AlertmanagerToBigQueryConfig struct {
	BigQuery BigQueryConfig    `yaml:"big_query"`
	Server   ServerConfig      `yaml:"server"`
	LabelMap map[string]string `yaml:"label_map"`
}

func LoadConfigFile(path string) AlertmanagerToBigQueryConfig {
	config := AlertmanagerToBigQueryConfig{}
	config.BigQuery.CredentialsJSON = os.Getenv("AMTOBQ_BIGQUERY_CREDENTIALS_JSON")

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	return config
}
