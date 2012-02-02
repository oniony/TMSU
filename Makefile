SHELL=/bin/sh
VER=0.0.5
HGREV=$(shell hg id)

SRC_DIR=src/main
BIN_DIR=bin
DIST_DIR=tmsu-$(VER)
INSTALL_DIR=/usr/bin

BIN_FILE=tmsu
VER_FILE=version.gen.go
DIST_FILE=tmsu-$(VER).tgz

all: clean generate compile dist test

clean:
	### Clean ###
	pushd ${SRC_DIR}; gomake clean; popd
	rm -f $(SRC_DIR)/$(VER_FILE)
	rm -Rf $(BIN_DIR)
	rm -Rf $(DIST_DIR)
	rm -f $(DIST_FILE)

generate:
	### Generate ###
	echo "package main; var version = \"$(VER) ($(HGREV))\"" >$(SRC_DIR)/$(VER_FILE)

compile: generate
	### Compile ###
	pushd $(SRC_DIR); gomake; popd
	@mkdir -p $(BIN_DIR)
	cp $(SRC_DIR)/$(BIN_FILE) $(BIN_DIR)

test: compile
	### Test ###
	pushd $(SRC_DIR); gomake test; popd

dist: compile
	### Dist ###
	@mkdir -p $(DIST_DIR)
	cp -R $(BIN_DIR) $(DIST_DIR)
	cp LICENSE README $(DIST_DIR)
	tar czf $(DIST_FILE) $(DIST_DIR)
	rm -Rf $(DIST_DIR)

install:
	### Install ###
	cp $(BIN_DIR)/$(BIN_FILE) $(INSTALL_DIR)

uninstall:
	### Uninstall ###
	rm $(INSTALL_DIR)/$(BIN_NAME)
