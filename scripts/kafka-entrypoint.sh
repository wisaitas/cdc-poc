#!/bin/bash

# Start Kafka in background
/etc/kafka/docker/run &

KAFKA_PID=$!
KAFKA_BIN="/opt/kafka/bin"
BOOTSTRAP_SERVER="localhost:9092"
TOPICS_FILE="/topics.txt"

# Wait for Kafka to be ready
echo "Waiting for Kafka to be ready..."
until $KAFKA_BIN/kafka-broker-api-versions.sh --bootstrap-server $BOOTSTRAP_SERVER > /dev/null 2>&1; do
    sleep 2
done

echo "Kafka is ready. Creating topics..."

# Read topics from file and create them
while IFS= read -r line || [[ -n "$line" ]]; do
  # Skip empty lines and comments
  [[ -z "$line" || "$line" =~ ^# ]] && continue
  
  # Parse: topic_name|partitions|replication_factor|cleanup_policy
  IFS='|' read -r topic partitions replication cleanup_policy <<< "$line"
  
  # Default values
  partitions=${partitions:-1}
  replication=${replication:-1}
  cleanup_policy=${cleanup_policy:-delete}
  
  echo "Creating topic: $topic (partitions: $partitions, replication: $replication, cleanup.policy: $cleanup_policy)"
  $KAFKA_BIN/kafka-topics.sh --bootstrap-server $BOOTSTRAP_SERVER \
    --create \
    --if-not-exists \
    --topic "$topic" \
    --partitions "$partitions" \
    --replication-factor "$replication" \
    --config cleanup.policy="$cleanup_policy"
done < "$TOPICS_FILE"

echo "All topics created!"
$KAFKA_BIN/kafka-topics.sh --bootstrap-server $BOOTSTRAP_SERVER --list

# Wait for Kafka process
wait $KAFKA_PID