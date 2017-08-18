all:
	go build -o uvdt-tracker tracker.go
	go build -o uvdt-node node.go

run:
	go run tracker.go
	go run node.go
