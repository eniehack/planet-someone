FLAGS =
GO = go
BINARIES = planetctl picker
BINDIR = dist
REVISION = $(shell git rev-parse HEAD)
SRC = $(shell find . -type f -name '*.go' -print)
.PHONY: clean pre-build

all: pre-build $(BINARIES) hb

pre-build:
	mkdir -p ./$(BINDIR)
	cp -r ./db ./$(BINDIR)/
	cp -r ./template ./$(BINDIR)/

hb: pre-build $(SRC)
	$(GO) $(FLAGS) build -ldflags="-X main.GitRevision=$(REVISION)" -o ./$(BINDIR)/$@ ./cmd/$@/main.go

$(BINARIES): pre-build $(SRC) 
	$(GO) $(FLAGS) build -o ./$(BINDIR)/$@ ./cmd/$@/main.go

clean:
	rm -rf ./$(BINDIR)