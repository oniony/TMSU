SHELL=/bin/sh
VER=0.0.4
HGREV=$(shell hg id)

SRC_DIR=src/main
BIN_DIR=bin
DIST_DIR=tmsu-$(VER)
INSTALL_DIR=/usr/bin

BIN_FILE=tmsu
VER_FILE=version.gen.go
DIST_FILE=tmsu-$(VER).tgz

all: clean generate compile

compile: generate
	pushd $(SRC_DIR); gomake; popd
	mkdir -p $(BIN_DIR)
	cp $(SRC_DIR)/$(BIN_FILE) $(BIN_DIR)

generate:
	echo "package main; var version = \"$(VER) ($(HGREV))\"" >$(SRC_DIR)/$(VER_FILE)

dist: compile
	mkdir -p $(DIST_DIR)
	cp -R $(BIN_DIR) $(DIST_DIR)
	cp LICENSE README $(DIST_DIR)
	tar czf $(DIST_FILE) $(DIST_DIR)
	rm -Rf $(DIST_DIR)

clean:
	pushd ${SRC_DIR}; gomake clean; popd
	rm -f $(SRC_DIR)/$(VER_FILE)
	rm -Rf $(BIN_DIR)
	rm -Rf $(DIST_DIR)
	rm -Rf $(DIST_FILE)

install:
	cp $(BIN_DIR)/$(BIN_FILE) $(INSTALL_DIR)

uninstall:
	rm $(INSTALL_DIR)/$(BIN_NAME)
