build:
	rm -rf ./dist
	mkdir -p ./dist
	go build -o ./dist/planetctl ./bin/ctl/main.go
	go build -o ./dist/picker ./bin/picker/main.go
	go build -o ./dist/hb ./bin/html-builder/main.go
