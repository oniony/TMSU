# installation paths
INSTALL_DIR=/usr/bin
MOUNT_INSTALL_DIR=/usr/sbin
MAN_INSTALL_DIR=/usr/share/man
ZSH_COMP_INSTALL_DIR=/usr/share/zsh/site-functions

# other vars
VER=0.3.0
SHELL=/bin/sh
HGREV=$(shell hg id)
ARCH=$(shell uname -m)
DIST_FILE=tmsu-$(ARCH)-$(VER).tgz

export GOPATH:=$(GOPATH):$(PWD)

all: clean generate compile dist test

clean:
	go clean tmsu
	rm -f src/tmsu/common/version.gen.go
	rm -Rf bin
	rm -Rf dist
	rm -f $(DIST_FILE)

generate:
	echo "package common; var Version = \"$(VER) ($(HGREV))\"" >src/tmsu/common/version.gen.go

compile: generate
	@mkdir -p bin
	go build -o bin/tmsu tmsu

test: compile
	go test tmsu/...

dist: compile
	@mkdir -p dist
	cp -R bin dist
	cp README.md dist
	cp COPYING dist
	@mkdir -p dist/man
	gzip -kfc misc/man/tmsu.1 >dist/man/tmsu.1.gz
	@mkdir -p dist/misc/zsh
	cp misc/zsh/_tmsu dist/misc/zsh
	tar czf $(DIST_FILE) dist

install:
	@echo "Installing TMSU"
	@echo -n "    "
	cp bin/tmsu $(INSTALL_DIR)
	@echo "Installing 'mount' command support"
	@echo -n "    "
	cp sbin/mount.tmsu $(MOUNT_INSTALL_DIR)
	@echo "Installing man page"
	mkdir -p $(MAN_INSTALL_DIR)
	gzip -kfc misc/man/tmsu.1 >$(MAN_INSTALL_DIR)/tmsu.1.gz
	@echo "Installing Zsh completion"
	@echo -n "    "
	mkdir -p $(ZSH_COMP_INSTALL_DIR)
	@echo -n "    "
	cp misc/zsh/_tmsu $(ZSH_COMP_INSTALL_DIR)

uninstall:
	@echo "Uninstalling TMSU"
	@echo -n "    "
	rm $(INSTALL_DIR)/tmsu
	@echo "Uninstalling mount support"
	@echo -n "    "
	rm $(MOUNT_INSTALL_DIR)/mount.tmsu
	@echo "Uninstalling man page"
	rm $(MAN_INSTALL_DIR)/tmsu.1.gz
	@echo "Uninstalling Zsh completion"
	@echo -n "    "
	rm $(ZSH_COMP_INSTALL_DIR)/_tmsu
