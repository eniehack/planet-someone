FLAGS =
GO = go
BINARIES = planetctl picker hb
BINDIR = dist
SRC = $(shell find . -type f -name '*.go' -print)
.PHONY: clean pre-build

all: pre-build $(BINARIES)

pre-build:
	mkdir -p ./$(BINDIR)
	cp -r ./db ./$(BINDIR)/
	cp -r ./template ./$(BINDIR)/

$(BINARIES): pre-build $(SRC) 
	$(GO) $(FLAGS) build -o ./$(BINDIR)/$@ ./cmd/$@/main.go

clean:
	rm -rf ./$(BINDIR)