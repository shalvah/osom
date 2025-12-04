package config

import mqtt "github.com/eclipse/paho.mqtt.golang"

var MQTTClient mqtt.Client

func InitMQTT() {
	MQTTClient = mqtt.NewClient(
		mqtt.NewClientOptions().
			AddBroker(Config.MQTTUrl).
			SetClientID("osom-server").
			SetUsername(Config.MQTTUsername).
			SetPassword(Config.MQTTPassword),
	)
}
