build:
	rm -rf ./dist
	mkdir -p ./dist
	go build -o ./dist/planetctl ./bin/planetctl/main.go
	go build -o ./dist/picker ./bin/picker/main.go
	go build -o ./dist/hb ./bin/hb/main.go

