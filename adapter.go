package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2api "github.com/influxdata/influxdb-client-go/v2/api"
)

const (
	influxDBAddress = "http://influxdb:8086"
	influxDBToken   = "9b_cEv2zuKpXpOu-uLE0X9jOsfP3JcWgMT9_l3sPT9CRED56zM1jjJRPZIujZvpH3TZaOoWMz4QhkL_bQ73amg=="
	influxDBOrg     = "SPRC"
	influxDBBucket  = "influxBucket"
	mqttAddress     = "mosquitto:1883"
	mqttTopic       = "#"
	mqttRegex       = "([^/]+)/([^/]+)"
	mqttClientID    = "MQTTInfluxDBBridge"
)

// SensorData struct defines the structure of data received from sensors
type SensorData struct {
	Location    string
	Measurement string
	Value       float64
	Timestamp   string
	Station     string
}

// Global variables for InfluxDB client and write API
var (
	influxDBClient influxdb2.Client
	writeAPI       influxdb2api.WriteAPI
)

// log function writes a debug message if debugging is enabled
func log(message string) {
	if os.Getenv("DEBUG_DATA_FLOW") == "true" {
		fmt.Println(message)
	}
}

// onMessageReceived handles incoming MQTT messages
func onMessageReceived(client mqtt.Client, message mqtt.Message) {
	var jsonPayload map[string]interface{}
	json.Unmarshal(message.Payload(), &jsonPayload) // Parse the JSON payload

	// Get and log the timestamp of the received message
	timestampStr, _ := jsonPayload["timestamp"].(string)
	now := time.Now().UTC().Format("2006-01-02T15:04:05")
	log(fmt.Sprintf("%s Received a message by topic %s with timestamp %s", now, message.Topic(), timestampStr))

	// Process each key-value pair in the JSON payload
	for key, value := range jsonPayload {
		if numVal, ok := value.(float64); ok { // Check if the value is a float64
			sensorData := parseMQTTMessage(message.Topic(), key, numVal, timestampStr)
			if sensorData != nil { // If parsing is successful
				sendSensorDataToInfluxDB(*sensorData) // Send the data to InfluxDB
				log(fmt.Sprintf("%s %s.%s %v", now, sensorData.Location, sensorData.Measurement, sensorData.Value))
			}
		}
	}
}

// parseMQTTMessage parses the topic and data from an MQTT message
func parseMQTTMessage(topic, measurement string, value float64, timestampStr string) *SensorData {
	r := regexp.MustCompile(mqttRegex)     // Compile the regex pattern
	matches := r.FindStringSubmatch(topic) // Match the topic against the regex
	if len(matches) > 2 {
		return &SensorData{ // Create and return a SensorData instance
			Location:    matches[1],
			Measurement: matches[2] + "." + measurement,
			Value:       value,
			Timestamp:   timestampStr,
			Station:     matches[2],
		}
	}
	return nil
}

// sendSensorDataToInfluxDB sends sensor data to InfluxDB
func sendSensorDataToInfluxDB(data SensorData) {
	timestamp := time.Now()
	if data.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, data.Timestamp); err == nil {
			timestamp = t // Use the parsed timestamp if available
		}
	}

	// Create a new point and write it to InfluxDB
	p := influxdb2.NewPoint(data.Measurement,
		map[string]string{"location": data.Location, "station": data.Station},
		map[string]interface{}{"value": data.Value},
		timestamp,
	)
	writeAPI.WritePoint(p)
	writeAPI.Flush() // Ensure all pending writes are sent to the database
}

// main sets up the MQTT client and connects to the MQTT broker
func main() {
	influxDBClient = influxdb2.NewClientWithOptions(influxDBAddress, influxDBToken, influxdb2.DefaultOptions().SetBatchSize(20))
	writeAPI = influxDBClient.WriteAPI(influxDBOrg, influxDBBucket)

	opts := mqtt.NewClientOptions().AddBroker(mqttAddress).SetClientID(mqttClientID)
	opts.SetDefaultPublishHandler(onMessageReceived)
	opts.OnConnect = func(c mqtt.Client) {
		c.Subscribe(mqttTopic, 1, nil) // Subscribe to the MQTT topic
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error()) // Handle connection errors
	}

	for {
		time.Sleep(1 * time.Second) // Sleep to limit resource usage
	}
}
