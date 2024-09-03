FLAGS =
GO = go
.PHONY: clean pre-build

all: pre-build planetctl picker hb

pre-build:
	mkdir -p ./dist
	cp -r ./db ./dist/
	cp -r ./template ./dist/

planetctl: pre-build ./cmd/planetctl/main.go 
	$(GO) $(FLAGS) build -o ./dist/planetctl ./cmd/planetctl/main.go

picker: pre-build ./cmd/picker/main.go
	$(GO) $(FLAGS) build -o ./dist/picker ./cmd/picker/main.go

hb: pre-build ./cmd/hb/main.go
	$(GO) $(FLAGS) build -o ./dist/hb ./cmd/hb/main.go

clean:
	rm -rf ./dist