SHELL=/bin/sh
VER=0.0.3

all: clean generate compile

compile: generate
	cd src/tmsu; gomake
	mkdir -p bin
	cp src/tmsu/tmsu bin

generate:
	echo "package main; var version = \"$(VER)\"" >src/tmsu/version.go

dist: compile
	mkdir -p tmsu-$(VER)
	cp -R bin tmsu-$(VER)
	cp LICENSE README tmsu-$(VER)
	tar czf tmsu-$(VER).tgz tmsu-$(VER)
	rm -Rf tmsu-$(VER)

clean:
	rm -f src/tmsu/version.go
	rm -Rf bin
	rm -Rf tmsu-$(VER)
	rm -Rf tmsu-$(VER).tgz
