all: build

.PHONY: build

build: bin/server bin/client

clean:
	-rm -rf bin/

bin/server: server.go
	gom build -o $@ server.go

bin/client: client.go
	gom build -o $@ client.go

install-dependencies:
  gom install

release:
	git pull
	make install-dependencies
	make clean
	make
