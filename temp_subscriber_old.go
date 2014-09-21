package main

import (
	"fmt"
	"os"
	"time"
	"strings"
	"strconv"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"labix.org/v2/mgo"
)

const (
	Broker string = "tcp://127.0.0.1:1883"
	Topic  string = "/temperature/0"
	Mongod string = "127.0.0.1"
)

type Temperature struct {
	Temp float64
	Pressure float64
	Altitude float64
	Datetime string
}

func main() {
	for{
		fmt.Println("------ check temp -----")
		opts := MQTT.NewClientOptions()
		opts.SetBroker(Broker)
		opts.SetTraceLevel(MQTT.Off)
		opts.SetCleanSession(true)
		opts.SetClientId("subuser")

		num_received := 0
		choke := make(chan [2]string)
		opts.SetDefaultPublishHandler(func(client *MQTT.MqttClient, msg MQTT.Message) {
			choke <- [2]string{msg.Topic(), string(msg.Payload())}
		})

		client := MQTT.NewClient(opts)
		_, err := client.Start()
		if err != nil {
			panic(err)
		}

		filter, e := MQTT.NewTopicFilter(Topic, byte(0))
		if e != nil {
			fmt.Println(e)
			os.Exit(1)
		}

		client.StartSubscription(nil, filter)

		for num_received < 1 {
			incoming := <-choke
			res := splitData(incoming[1])
			setMongo(res[0], res[1], res[2])
			fmt.Printf("RECEIVED TOPIC %s MESSAGE: /%f/%f/%f\n", incoming[0], res[0], res[1], res[2])
			num_received++
		}

		client.Disconnect(250)

		fmt.Println("------ check temp end -----")
	}

}

func splitData(str string) []float64{
	strs := strings.Split(str, "/")
	var res []float64 = make([]float64, 3)

	tmp := strings.Split(strs[0], ":")
	temp := strings.Trim(tmp[1], " ")
	res[0], _ = strconv.ParseFloat(temp, 64)

	press := strings.Split(strs[1], ":")
	prs := strings.Trim(press[1], " ")
	res[1], _ = strconv.ParseFloat(prs, 64)

	altitude := strings.Split(strs[2], ":")
	altmp := strings.Trim(altitude[1], " ")

	// TODO なぜか改行が入ってるのでpublisherから改行を捨てる
	altmp = strings.Trim(altmp, "\n")


	a, err := strconv.ParseFloat(altmp, 64)
	if err != nil{
		fmt.Println("parse")
		panic(err)
	}
	res[2] = a
	return res
}


func setMongo(tmp, prs, alti float64){
	sess, err := mgo.Dial(Mongod)
	if err != nil{
		panic(err)
	}
	defer sess.Close()

	t := time.Now()
	now := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
	sess.SetMode(mgo.Monotonic, true)
	c := sess.DB("test").C("temperature")
	err = c.Insert(&Temperature{tmp, prs, alti, now})
	if err != nil{
		panic(err)
	}

}

