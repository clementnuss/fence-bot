package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	client       mqtt.Client
	stat         shellyStatus
	shellyPrefix string
)

type shellyStatus struct {
	InputStatus  shellyInputStatus
	SwitchStatus shellySwitchStatus
}

type shellyInputStatus struct {
	Id         int  `json:"id"`
	State      bool `json:"state"`
	LastUpdate time.Time
}
type shellySwitchStatus struct {
	Output       bool    `json:"output"`
	Voltage      float64 `json:"voltage"`
	Current      float64 `json:"current"`
	AveragePower float64 `json:"apower"`
	LastUpdate   time.Time
}

func handler(client mqtt.Client, message mqtt.Message) {
	switch message.Topic() {
	case shellyPrefix + "status/input:0":
		err := json.Unmarshal(message.Payload(), &stat.InputStatus)
		if err == nil {
			stat.InputStatus.LastUpdate = time.Now()
		}
	case shellyPrefix + "status/switch:0":
		err := json.Unmarshal(message.Payload(), &stat.SwitchStatus)
		if err == nil {
			stat.SwitchStatus.LastUpdate = time.Now()
		}
	}
}

func statusUpdate() {
	client.Publish(shellyPrefix+"command", 1, false, "status_update")
}

func toggleOn() {
	client.Publish(shellyPrefix+"command/switch:0", 0, false, "on")
}

func toggleOff() {
	client.Publish(shellyPrefix+"command/switch:0", 0, false, "off")
}

func mqttInit() error {
	host, err := os.Hostname()
	if err != nil {
		return err
	}

	shellyPrefix = os.Getenv("SHELLY_PREFIX")
	mqttServer := os.Getenv("MQTT_SERVER")

	// mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	// mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	// mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	opts := mqtt.NewClientOptions().
		AddBroker(mqttServer).
		// SetTLSConfig(&tls.Config{InsecureSkipVerify: true}).
		SetClientID("fencebot-" + host).
		SetOnConnectHandler(func(client mqtt.Client) {
			topic := shellyPrefix + "#"
			token := client.Subscribe(
				topic,
				0, /* minimal QoS level zero: at most once, best-effort delivery */
				handler)
			if !token.WaitTimeout(5*time.Second) && token.Error() != nil {
				log.Fatal(token.Error())
			}
			log.Printf("subscribed to %q", topic)
		}).
		SetConnectRetry(true)

	client = mqtt.NewClient(opts)
	// mqttMessageHandler.client = client
	if token := client.Connect(); token.WaitTimeout(5*time.Second) && token.Error() != nil {
		// This can indeed fail, e.g. if the broker DNS is not resolvable.
		return fmt.Errorf("MQTT connection failed: %v", token.Error())
	}
	log.Printf("MQTT subscription established")

	return nil
}
