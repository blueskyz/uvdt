all:
	mkdir -pv bin
	go build -o bin/uvdt-tracker tracker.go
	go build -o bin/uvdt-node node.go
	go build -o bin/uvdt-node-tool node-tool.go


# run:
	# go run tracker.go
	# go run node.go
	# go run node-tool.go


clean:
	rm -r -f -v bin/uvdt-tracker bin/uvdt-node bin/uvdt-node-tool
