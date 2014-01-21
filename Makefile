# installation paths
INSTALL_DIR=/usr/bin
MOUNT_INSTALL_DIR=/usr/sbin
MAN_INSTALL_DIR=/usr/share/man/man1
ZSH_COMP_INSTALL_DIR=/usr/share/zsh/site-functions

# other vars
VER=0.4.0
SHELL=/bin/sh
ARCH=$(shell uname -m)
DIST_NAME=tmsu-$(ARCH)-$(VER)
DIST_DIR=$(DIST_NAME)
DIST_FILE=$(DIST_NAME).tgz

export GOPATH:=$(GOPATH):$(PWD)

all: clean generate compile dist test

clean:
	go clean tmsu
	rm -f src/tmsu/common/version.gen.go
	rm -Rf bin
	rm -Rf $(DIST_DIR)
	rm -f $(DIST_FILE)

generate:
	echo "package common; var Version = \"$(VER)\"" >src/tmsu/common/version.gen.go

compile: generate
	@mkdir -p bin
	go build -o bin/tmsu tmsu

test: compile
	go test tmsu/...

dist: compile
	@mkdir -p $(DIST_DIR)
	cp -R bin $(DIST_DIR)
	cp README.md $(DIST_DIR)
	cp COPYING $(DIST_DIR)
	@mkdir -p $(DIST_DIR)/bin
	cp misc/bin/mount.tmsu $(DIST_DIR)/bin/
	@mkdir -p $(DIST_DIR)/man
	gzip -kfc misc/man/tmsu.1 >$(DIST_DIR)/man/tmsu.1.gz
	@mkdir -p $(DIST_DIR)/misc/zsh
	cp misc/zsh/_tmsu $(DIST_DIR)/misc/zsh/
	tar czf $(DIST_FILE) $(DIST_DIR)

install:
	@echo "* Installing TMSU"
	cp bin/tmsu $(INSTALL_DIR)
	@echo "* Installing 'mount' command support"
	cp misc/bin/mount.tmsu $(MOUNT_INSTALL_DIR)
	@echo "* Installing man page"
	mkdir -p $(MAN_INSTALL_DIR)
	gzip -kfc misc/man/tmsu.1 >$(MAN_INSTALL_DIR)/tmsu.1.gz
	@echo "* Installing Zsh completion"
	mkdir -p $(ZSH_COMP_INSTALL_DIR)
	cp misc/zsh/_tmsu $(ZSH_COMP_INSTALL_DIR)

uninstall:
	@echo "* Uninstalling TMSU"
	rm $(INSTALL_DIR)/tmsu
	@echo "* Uninstalling mount support"
	rm $(MOUNT_INSTALL_DIR)/mount.tmsu
	@echo "* Uninstalling man page"
	rm $(MAN_INSTALL_DIR)/tmsu.1.gz
	@echo "* Uninstalling Zsh completion"
	rm $(ZSH_COMP_INSTALL_DIR)/_tmsu
