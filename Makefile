VER=0.1.1
SHELL=/bin/sh
HGREV=$(shell hg id)

SRC_DIR=src/tmsu
BIN_DIR=bin
DIST_DIR=tmsu-$(VER)
INSTALL_DIR=/usr/bin

ZSH_COMP=misc/zsh/_tmsu
ZSH_COMP_INSTALL_DIR=/usr/share/zsh/site-functions

BIN_FILE=tmsu
VER_FILE=version.gen.go
ARCH=$(shell uname -m)
DIST_FILE=tmsu-$(ARCH)-$(VER).tgz

export GOPATH=$(PWD)

all: clean generate compile dist test

clean:
	go clean tmsu
	rm -f $(SRC_DIR)/common/$(VER_FILE)
	rm -Rf $(BIN_DIR)
	rm -Rf $(DIST_DIR)
	rm -f $(DIST_FILE)

generate:
	echo "package common; var Version = \"$(VER) ($(HGREV))\"" >$(SRC_DIR)/common/$(VER_FILE)

compile: generate
	go build -o $(BIN_FILE) tmsu
	@mkdir -p $(BIN_DIR)
	mv $(BIN_FILE) $(BIN_DIR)

test: compile
	go test tmsu/...

dist: compile
	@mkdir -p $(DIST_DIR)
	cp -R $(BIN_DIR) $(DIST_DIR)
	cp README.md $(DIST_DIR)
	cp COPYING $(DIST_DIR)
	tar czf $(DIST_FILE) $(DIST_DIR)
	rm -Rf $(DIST_DIR)

install:
	cp $(BIN_DIR)/$(BIN_FILE) $(INSTALL_DIR)
	@mkdir -p $(ZSH_COMP_INSTALL_DIR)
	cp $(ZSH_COMP) $(ZSH_COMP_INSTALL_DIR)

uninstall:
	rm $(INSTALL_DIR)/$(BIN_NAME)
