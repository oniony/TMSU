VER=0.1.3

compile: clean version
	cd src/tmsu; gomake
	mkdir bin
	cp src/tmsu/tmsu bin

version:
	echo "package main; var version = \"$(VER)\"" >src/tmsu/version.go

package: compile
	mkdir tmsu-$(VER)
	cp -R bin tmsu-$(VER)
	cp LICENSE README tmsu-$(VER)
	tar czf tmsu-$(VER).tgz tmsu-$(VER)
	rm -Rf tmsu-$(VER)

clean:
	rm -f src/tmsu/version.go
	rm -Rf bin
	rm -f tmsu-$(VER).tgz
