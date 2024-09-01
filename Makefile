FLAGS =
GO = go
.PHONY: clean

all: pre-build planetctl picker hb

pre-build:
	mkdir -p ./dist

planetctl: ./bin/planetctl/main.go
	$(GO) $(FLAGS) build -o ./dist/planetctl ./bin/planetctl/main.go

picker: ./bin/picker/main.go
	$(GO) $(FLAGS) build -o ./dist/picker ./bin/picker/main.go

hb: ./bin/hb/main.go
	$(GO) $(FLAGS) build -o ./dist/hb ./bin/hb/main.go

clean:
	rm -rf ./dist