package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	WeatherapiKey string `mapstructure:"WEATHER_API_KEY"`
}

func LoadConfig() (*Config, error) {
	// 1. Aponta para o diretório onde o .env está (neste caso, o diretório atual).
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// 2. Habilita a leitura automática de variáveis de ambiente do SO.
	viper.AutomaticEnv()

	// 3. Tenta ler o arquivo de configuração .env.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Se o erro for diferente de "arquivo não encontrado", retorne o erro.
			return nil, fmt.Errorf("erro ao ler o arquivo de configuração: %w", err)
		}
		// Se o arquivo .env não existe, não há problema, continue.
	}

	// 4. Faz o "Unmarshal" dos valores encontrados para a struct Config.
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("erro ao fazer unmarshal da configuração: %w", err)
	}

	return &config, nil
}
