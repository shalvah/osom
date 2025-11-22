package config

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	DefaultLatLong        string
	NextBikesApiUrl       string
	VrnApiUrl             string
	MQTTUrl               string
	MQTTUsername          string
	MQTTPassword          string
	AvailabilityMQTTTopic string
}

var Config AppConfig

func Init() {
	viper.AutomaticEnv()

	// read env values explicitly into the struct (AutomaticEnv doesnt work with Unmarshal)
	Config.DefaultLatLong = viper.GetString("DEFAULT_LATLONG")
	Config.NextBikesApiUrl = viper.GetString("NEXTBIKES_API_URL")
	Config.VrnApiUrl = viper.GetString("VRN_API_URL")
	Config.MQTTUrl = viper.GetString("MQTT_URL")
	Config.MQTTUsername = viper.GetString("MQTT_USERNAME")
	Config.MQTTPassword = viper.GetString("MQTT_PASSWORD")
	Config.AvailabilityMQTTTopic = viper.GetString("AVAILABILITY_MQTT_TOPIC")
}
