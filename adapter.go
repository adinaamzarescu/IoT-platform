package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	// MQTT package for handling messages and communication over MQTT protocol.
	mqtt "github.com/eclipse/paho.mqtt.golang"
	// InfluxDB packages for connecting to an InfluxDB and managing data.
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2api "github.com/influxdata/influxdb-client-go/v2/api"
)

// Constants for configuring the InfluxDB and MQTT connections.
const (
	influxDBAddress = "http://influxdb:8086"                                                                     // InfluxDB server address
	influxDBToken   = "9b_cEv2zuKpXpOu-uLE0X9jOsfP3JcWgMT9_l3sPT9CRED56zM1jjJRPZIujZvpH3TZaOoWMz4QhkL_bQ73amg==" // InfluxDB's authentication token
	influxDBOrg     = "SPRC"                                                                                     // Organization for InfluxDB
	influxDBBucket  = "influxBucket"                                                                             // Data bucket in InfluxDB

	mqttAddress  = "mosquitto:1883"     // MQTT server address
	mqttTopic    = "#"                  // MQTT topic to subscribe to
	mqttRegex    = "([^/]+)/([^/]+)"    // Regex to parse MQTT topics
	mqttClientID = "MQTTInfluxDBBridge" // MQTT client identifier
)

// SensorData defines the structure for storing sensor data.
type SensorData struct {
	Location    string  // Location of the sensor
	Measurement string  // Measurement type
	Value       float64 // Measurement value
	Timestamp   string  // Time of the measurement
	Station     string  // Station identifier
}

// Global variables for InfluxDB client and API.
var influxDBClient influxdb2.Client
var writeAPI influxdb2api.WriteAPI

// log function for conditional logging based on the environment variable.
func log(message string) {
	if os.Getenv("DEBUG_DATA_FLOW") == "true" {
		fmt.Println(message)
	}
}

// isFloat checks if a string can be converted to a float64.
func isFloat(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

// isInt checks if a string can be precisely converted to an int64.
func isInt(value string) bool {
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		i := int64(f)
		return float64(i) == f
	}
	return false
}

// onMessageReceived is the callback for handling incoming MQTT messages.
func onMessageReceived(client mqtt.Client, message mqtt.Message) {
	now := time.Now().UTC().Format("2006-01-02T15:04:05")        // Current time in UTC
	log(now + " Received a message by topic " + message.Topic()) // Log the received message

	var jsonPayload map[string]interface{}
	json.Unmarshal(message.Payload(), &jsonPayload) // Decode JSON payload
	timestampStr := ""
	for key, value := range jsonPayload {
		if strVal, ok := value.(string); ok && key == "timestamp" {
			timestampStr = strVal
		}
	}

	if timestampStr != "" {
		log(now + " Data timestamp is: " + timestampStr)
	} else {
		log(now + " Data timestamp is: NOW")
	}

	// Process each key-value pair in the JSON payload.
	for key, value := range jsonPayload {
		strValue := fmt.Sprintf("%v", value)
		if isInt(strValue) || isFloat(strValue) {
			val, _ := strconv.ParseFloat(strValue, 64)
			sensorData := parseMQTTMessage(message.Topic(), key, val, timestampStr)
			if sensorData != nil {
				sendSensorDataToInfluxDB(*sensorData) // Send valid sensor data to InfluxDB
				log(now + " " + sensorData.Location + "." + sensorData.Measurement + " " + strconv.FormatFloat(sensorData.Value, 'f', -1, 64))
			}
		}
	}
}

// parseMQTTMessage parses the MQTT topic to extract sensor data.
func parseMQTTMessage(topic, measurement string, value float64, timestampStr string) *SensorData {
	r := regexp.MustCompile(mqttRegex) // Compile the regex pattern
	matches := r.FindStringSubmatch(topic)
	if matches != nil && len(matches) > 2 {
		location := matches[1]
		station := matches[2]
		measurement := station + "." + measurement
		return &SensorData{Location: location, Measurement: measurement, Value: value, Timestamp: timestampStr, Station: station}
	}
	return nil
}

// sendSensorDataToInfluxDB writes the sensor data to InfluxDB.
func sendSensorDataToInfluxDB(data SensorData) {
	timestamp := time.Now()
	if data.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, data.Timestamp); err == nil {
			timestamp = t
		}
	}

	p := influxdb2.NewPoint(data.Measurement,
		map[string]string{
			"location": data.Location,
			"station":  data.Station,
		},
		map[string]interface{}{
			"value": data.Value,
		},
		timestamp,
	)
	writeAPI.WritePoint(p) // Write data to InfluxDB
	writeAPI.Flush()       // Ensure all pending data is sent to InfluxDB
}

// main function initializes connections and runs the MQTT client.
func main() {
	influxDBClient = influxdb2.NewClientWithOptions(influxDBAddress, influxDBToken, influxdb2.DefaultOptions().SetBatchSize(20))
	writeAPI = influxDBClient.WriteAPI(influxDBOrg, influxDBBucket)

	opts := mqtt.NewClientOptions().AddBroker(mqttAddress).SetClientID(mqttClientID)
	opts.SetDefaultPublishHandler(onMessageReceived)
	opts.OnConnect = func(c mqtt.Client) {
		c.Subscribe(mqttTopic, 1, nil) // Subscribe to MQTT topic upon connecting
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error()) // Handle connection error
	}

	for {
		time.Sleep(1 * time.Second) // Prevent exit and allow continuous operation
	}
}
