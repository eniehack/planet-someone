FLAGS =
GO = go
.PHONY: clean pre-build

all: pre-build planetctl picker hb

pre-build:
	mkdir -p ./dist
	cp -r ./db ./dist/
	cp -r ./template ./dist/

planetctl: pre-build ./bin/planetctl/main.go 
	$(GO) $(FLAGS) build -o ./dist/planetctl ./bin/planetctl/main.go

picker: pre-build ./bin/picker/main.go
	$(GO) $(FLAGS) build -o ./dist/picker ./bin/picker/main.go

hb: pre-build ./bin/hb/main.go
	$(GO) $(FLAGS) build -o ./dist/hb ./bin/hb/main.go

clean:
	rm -rf ./dist