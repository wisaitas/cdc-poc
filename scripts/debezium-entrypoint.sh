#!/bin/bash

# Start Kafka Connect in background
/docker-entrypoint.sh start &

# Wait for Kafka Connect to be ready
echo "Waiting for Kafka Connect to be ready..."
until curl -s http://localhost:8083/connectors > /dev/null 2>&1; do
    sleep 2
done

echo "Kafka Connect is ready. Registering connector..."

# Register the PostgreSQL connector
curl -X POST http://localhost:8083/connectors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "postgres-connector",
    "config": {
      "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
      "database.hostname": "postgres",
      "database.port": "5432",
      "database.user": "postgres",
      "database.password": "postgres",
      "database.dbname": "mydb",
      "topic.prefix": "postgres-connector",
      "table.include.list": "public.messages",
      "plugin.name": "pgoutput",
      "publication.autocreate.mode": "filtered",
      "slot.name": "debezium_slot",
      "schema.history.internal.kafka.bootstrap.servers": "kafka:29092",
      "schema.history.internal.kafka.topic": "schema-changes.mydb"
    }
  }'

echo ""
echo "Connector registered!"

# Wait for Kafka Connect process
wait