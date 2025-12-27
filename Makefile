.PHONY: runproducer runconsumer

runproducer:
	go run cmd/producer/main.go

runconsumer:
	go run cmd/consumer/main.go