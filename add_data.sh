#!/bin/bash

# Loop to execute the mosquitto_pub command 15 times with different values
for i in {1..15}
do
  # Generate random values for BAT, HUMID, and TMP
  BAT=$((50 + RANDOM % 51))  # Random battery level between 50 and 100
  HUMID=$((20 + RANDOM % 81))  # Random humidity level between 20 and 100
  TMP=$((10 + RANDOM % 31))  # Random temperature between 10 and 40 degrees Celsius

  # Calculate a random timestamp from 6 hours ago to now
  SECONDS_PAST=$((RANDOM % 21600))  # 21600 seconds in 6 hours
  RANDOM_TIMESTAMP=$(date --iso-8601=seconds -d "-$SECONDS_PAST seconds")

  # Build the JSON payload
  MSG="{\"BAT\": $BAT, \"HUMID\": $HUMID, \"TMP\": $TMP, \"PRJ\": \"SPRC\", \"status\": \"OK\", \"timestamp\": \"$RANDOM_TIMESTAMP\"}"

  # Publish the message to the MQTT topic
  mosquitto_pub -h localhost -t "UPB/RPi_1" -m "$MSG"

  sleep 1

  # Change values for the next message
  BAT=$((2 + RANDOM % 99))  # Random battery level between 2 and 100
  HUMID=$((20 + RANDOM % 81))
  TMP=$((10 + RANDOM % 31))

  # Calculate another random timestamp
  SECONDS_PAST=$((RANDOM % 21600))  # 21600 seconds in 6 hours
  RANDOM_TIMESTAMP=$(date --iso-8601=seconds -d "-$SECONDS_PAST seconds")

  MSG="{\"BAT\": $BAT, \"HUMID\": $HUMID, \"TMP\": $TMP, \"PRJ\": \"SPRC\", \"status\": \"OK\", \"timestamp\": \"$RANDOM_TIMESTAMP\"}"

  mosquitto_pub -h localhost -t "UBB/RPi_1" -m "$MSG"

  sleep 1
done
