package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"os"
)

func Load(configPath string, configName string, rawVal interface{}) *viper.Viper {

	fmt.Printf("load %s %s\n", configPath, configName)

	cfg := viper.New()
	cfg.AddConfigPath(configPath)
	cfg.SetConfigName(configName)

	if err := cfg.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := cfg.Unmarshal(rawVal); err != nil {
		panic(err)
	}

	return cfg
}

func PrintWithJSON(v any) {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		os.Stdout.WriteString("json marshal error: " + err.Error())
		return
	}
	os.Stdout.WriteString(string(b) + "\n")
}
