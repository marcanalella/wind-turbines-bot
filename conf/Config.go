package conf

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Address                string `json:"address"`
	Port                   string `json:"port"`
	Env                    string `json:"env"`
	AllCommand             string `json:"all_command"`
	TelegramApiBaseUrl     string `json:"telegram_api_base_url"`
	TelegramApiSendMessage string `json:"telegram_api_send_message"`
	TelegramTokenEnv       string `json:"telegram_token_env"`
	NorthernPowerApiEnv    string `json:"northern_power_api_env"`
	NorthernPowerApiEnv1   string `json:"northern_power_api_env_1"`
	NorthernPowerUser      string `json:"northern_power_user"`
	NorthernPowerPassword  string `json:"northern_power_password"`
}

func Load() (Config, error) {
	var config Config
	configFile, err := os.Open("config.json")
	defer func(configFile *os.File) {
		err := configFile.Close()
		if err != nil {
			log.Printf("could not decode json config %s\n", err.Error())
		}
	}(configFile)
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return config, err
}
