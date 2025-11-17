package config

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	DefaultLatLong  string
	NextBikesApiUrl string
	VrnApiUrl       string
}

var Config AppConfig

func Init() {
	viper.AutomaticEnv()

	// read env values explicitly into the struct (AutomaticEnv doesnt work with Unmarshal)
	Config.DefaultLatLong = viper.GetString("DEFAULT_LATLONG")
	Config.NextBikesApiUrl = viper.GetString("NEXTBIKES_API_URL")
	Config.VrnApiUrl = viper.GetString("VRN_API_URL")
}
