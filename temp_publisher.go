package main

import (
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
	"github.com/kidoman/embd/sensor/bmp180"

	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

const (
	Broker string = "tcp://127.0.0.1:1883"
	Topic  string = "/temperature/0"
)

func main() {
	//	broker := "tcp://127.0.0.1:1883"
	//	topic := "/temperature/0"
	temp, press, alti := getTemp()

	payload := fmt.Sprintf("temp:%.1f / press:%.1f / altitude: %.1f\n", temp, press, alti)

	opts := MQTT.NewClientOptions()
	opts.SetBroker(Broker)
	opts.SetTraceLevel(MQTT.Off)
	opts.SetCleanSession(true)
	opts.SetClientId("pubuser")
	client := MQTT.NewClient(opts)
	_, err := client.Start()
	gotareceipt := make(chan bool)
	if err != nil {
		panic(err)
	}
	fmt.Println("----- doing publish ----")
	receipt := client.Publish(MQTT.QoS(0), Topic, []byte(payload))
	go func() {
		<-receipt
		fmt.Println("  message delivered!")
		gotareceipt <- true
	}()

	for i := 0; i < 1; i++ {
		<-gotareceipt
	}

	client.Disconnect(250)
	fmt.Println("Sample Publisher Disconnected")
}

func getTemp() (float64, float64, float64) {
	err := embd.InitI2C()
	if err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)
	baro := bmp180.New(bus)
	defer baro.Close()

	temp, err := baro.Temperature()
	if err != nil {
		panic(err)
	}
	press, err := baro.Pressure()
	if err != nil {
		panic(err)
	}
	press64 := float64(press/100) + (float64(press%100) * 0.01)
	alti, err := baro.Altitude()
	if err != nil {
		panic(err)
	}

	return temp, press64, alti
}
