SHELL=/bin/sh
VER=0.0.5
HGREV=$(shell hg id)
export GOPATH=$(PWD)

SRC_DIR=src/tmsu
BIN_DIR=bin
DIST_DIR=tmsu-$(VER)
INSTALL_DIR=/usr/bin
ZSH_COMP=misc/zsh/_tmsu
ZSH_COMP_INSTALL_DIR=/usr/share/zsh/site-functions

BIN_FILE=tmsu
VER_FILE=version.gen.go
DIST_FILE=tmsu-$(VER).tgz

all: clean generate compile dist test

clean:
	### Clean ###
	pushd ${SRC_DIR}; go tool make clean; popd
	rm -f $(SRC_DIR)/core/$(VER_FILE)
	rm -Rf $(BIN_DIR)
	rm -Rf $(DIST_DIR)
	rm -f $(DIST_FILE)

generate:
	### Generate ###
	echo "package core; var Version = \"$(VER) ($(HGREV))\"" >$(SRC_DIR)/core/$(VER_FILE)

compile: generate
	### Compile ###
	go build -o $(BIN_FILE) tmsu/main
	@mkdir -p $(BIN_DIR)
	mv $(BIN_FILE) $(BIN_DIR)

test: compile
	### Test ###
	go test tmsu/...

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
	@mkdir -p $(ZSH_COMP_INSTALL_DIR)
	cp $(ZSH_COMP) $(ZSH_COMP_INSTALL_DIR)

uninstall:
	### Uninstall ###
	rm $(INSTALL_DIR)/$(BIN_NAME)
