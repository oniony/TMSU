SHELL=/bin/sh
VER=0.0.3

SRC_DIR=src/main
BIN_DIR=bin
DIST_DIR=tmsu-$(VER)
INSTALL_DIR=/usr/bin

BIN_FILE=tmsu
VER_FILE=version.gen.go
DIST_FILE=tmsu-$(VER).tgz

all: clean generate compile

compile: generate
	cd $(SRC_DIR); gomake
	mkdir -p $(BIN_DIR)
	cp $(SRC_DIR)/$(BIN_FILE) $(BIN_DIR)

generate:
	echo "package main; var version = \"$(VER)\"" >$(SRC_DIR)/$(VER_FILE)

dist: compile
	mkdir -p $(DIST_DIR)
	cp -R $(BIN_DIR) $(DIST_DIR)
	cp LICENSE README $(DIST_DIR)
	tar czf $(DIST_FILE) $(DIST_DIR)
	rm -Rf $(DIST_DIR)

clean:
	rm -f $(SRC_DIR)/$(VER_FILE)
	rm -f $(SRC_DIR)/$(BIN_FILE)
	rm -f $(SRC_DIR)/*.8
	rm -Rf $(BIN_DIR)
	rm -Rf $(DIST_DIR)
	rm -Rf $(DIST_FILE)

install:
	cp $(BIN_DIR)/$(BIN_FILE) $(INSTALL_DIR)

uninstall:
	rm $(INSTALL_DIR)/$(BIN_NAME)
