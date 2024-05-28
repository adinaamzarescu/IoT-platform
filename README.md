# IoT Platform

Docker-based system designed for IoT data collection and visualization. It uses MQTT for message transmission, InfluxDB for data storage, and Grafana for data visualization, orchestrated with Docker Swarm for scalable deployment.

## Components

- **Mosquitto**: An MQTT broker that handles messages between IoT devices and the server.
- **InfluxDB**: A time-series database optimized for the efficient storage and retrieval of time-stamped data.
- **Grafana**: A tool that provides visualization dashboards to monitor and analyze sensor data.
- **Adapter**: A Go application that bridges MQTT messages to InfluxDB, ensuring data from sensors is stored and ready for analysis.

## Adapter Application Details

The `adapter` application in Go acts as a middleware that captures data transmitted over MQTT, processes this data, and stores it in InfluxDB.

### Functionality

#### `main`
- **Purpose**: Initializes and manages connections to MQTT and InfluxDB, ensuring the application continuously processes incoming messages.
- **Details**: Configures and connects the MQTT client, subscribes to topics, and enters a loop to keep the service alive to handle incoming data.

#### `onMessageReceived`
- **Purpose**: Receives and handles incoming MQTT messages, extracting and processing data for storage.
- **Details**: Parses JSON payloads from MQTT messages, logs each message, and processes data entries to be stored in InfluxDB.

#### `parseMQTTMessage`
- **Purpose**: Parses the MQTT topic to extract location and sensor information, forming structured sensor data.
- **Details**: Uses a regular expression to dissect the topic, extracting relevant parts to create a `SensorData` object if the topic matches expected patterns.

#### `sendSensorDataToInfluxDB`
- **Purpose**: Sends structured sensor data to InfluxDB for storage.
- **Details**: Converts `SensorData` into database points, managing timestamps and ensuring data is committed to the database efficiently.

#### `log`
- **Purpose**: Provides a debugging output mechanism controlled by an environment variable.
- **Details**: Outputs messages to standard output if debugging is enabled, facilitating troubleshooting.

## `stack.yaml` Configuration

The `stack.yaml` file defines the deployment configuration for Docker Swarm.

### Network Configuration

- **Service Networks**: Each service (Mosquitto, InfluxDB, Grafana, and the Adapter) is configured to communicate over designated Docker overlay networks:
  - `mosquitto_adapter_network`: Connects the Mosquitto broker with the Adapter.
  - `influxdb_adapter_network`: Allows communication between InfluxDB and the Adapter.
  - `influxdb_grafana_network`: Facilitates connectivity between InfluxDB and Grafana for data visualization.
  
  These networks are defined as `overlay` to support Docker Swarm.

### Service Details

- **Mosquitto**: Configured with volumes for persistence and a custom configuration file, exposed on port 1883.
- **InfluxDB**: Set up with initialization parameters for the database, including default credentials and configurations, exposed on port 8086.
- **Grafana**: Configured with pre-provisioned data sources and dashboards, exposed on port 3000.
- **Adapter**: Custom Docker image that depends on both Mosquitto and InfluxDB for its operation.


### How to run

On Linux:

```
./run.sh
```

On Windows:

```
.\run.ps1
```

This script checks if Docker Swarm is active and if not, it initializes it. 
Then, it proceeds to deploy the defined stack using the configurations specified in stack.yaml.


Then on Linux/WSL run:

```
./add_data.sh
```

This script simulates sensor data input by publishing random values for battery level, humidity, and temperature to the MQTT broker. 
It runs a loop 15 times, each time generating a new set of random values along with a timestamp indicating the time of the data point. 
The script uses the mosquitto_pub command to publish each generated JSON payload to specific MQTT topics, which are then picked up by 
the adapter application for processing and storage in InfluxDB.

Each iteration of the loop performs the following actions:

1. Generates random sensor values for battery (BAT), humidity (HUMID), and temperature (TMP).
2. Calculates a random timestamp from the last 6 hours.
3. Constructs a JSON message with these values.
4. Publishes the message to an MQTT topic using mosquitto_pub.
5. Waits for a second before repeating the process with new values.

### Test the application

Then on Linux/WSL run(before running the add_data script):

```
mosquitto_sub -h localhost -t "#" -v
```

Then navigate to:

* [InfluxDB](http://localhost:8086/)
* [Grafana](http://localhost:3000/)

For authentication use: 

* InfluxDB
  - username = asistent
  - password = influxSPRC2023
 
* Grafana
  - username = asistent
  - password = grafanaSPRC2023

### Stop the application

```
docker service rm sprc_influxdb sprc_grafana sprc_adapter sprc_mosquitto
```

```
docker swarm leave --force
```

## Resoruces

1. https://github.com/eclipse/paho.mqtt.golang
2. github.com/influxdata/influxdb-client-go/v2
3. github.com/influxdata/influxdb-client-go/v2/api
4. https://mobylab.docs.crescdi.pub.ro/docs/softwareDevelopment/laboratory2/swarm
5. https://gitlab.com/mobylab-idp

