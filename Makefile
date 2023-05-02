all: ccc
	go build

ccc:
	go build

install: ccc
	sudo install ./ccc /usr/local/bin/ccc

clean:
	rm ccc
