.PHONY: init
init: topic.create.orders db.migrate.init
	@echo "Kafka, PostgreSQL"

# KAFKA
KAFKA_CONTAINER = l0-kafka-1
KAFKA_BROKER = localhost:9092

.PHONY: topic.create.orders
topic.create.orders:
	docker exec $(KAFKA_CONTAINER) \
		kafka-topics.sh --create \
		--topic orders \
		--bootstrap-server $(KAFKA_BROKER) \
		--partitions 1 \
		--replication-factor 1
	@echo "topic created"

# POSTGRES
POSTGRES_CONTAINER = l0-db-1
POSTGRES_DB = l0db
POSTGRES_USER = l0user
POSTGRES_PASSWORD = l0pass
SQL_FILE = migrations/init.sql


.PHONY: db.migrate.init
db.migrate.init:
	docker exec -i $(POSTGRES_CONTAINER) psql \
		--username=$(POSTGRES_USER) \
		--dbname=$(POSTGRES_DB) \
		< $(SQL_FILE)
	@echo "db tables created"

