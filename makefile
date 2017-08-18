all:
	mkdir -pv bin
	go build -o bin/uvdt-tracker tracker.go
	go build -o bin/uvdt-node node.go


run:
	go run tracker.go
	go run node.go


clean:
	rm -r -f -v bin/uvdt-tracker bin/uvdt-node
