# installation paths
INSTALL_DIR=$(DESTDIR)/usr/local/bin
MOUNT_INSTALL_DIR=$(DESTDIR)/usr/local/sbin
MAN_INSTALL_DIR=$(DESTDIR)/usr/local/share/man/man1
ZSH_COMP_INSTALL_DIR=$(DESTDIR)/usr/local/share/zsh/site-functions
BASH_COMP_INSTALL_DIR=$(DESTDIR)/usr/local/etc/bash_completion.d

# other vars
VER=$(shell grep -o "[0-9]\+\.[0-9]\+\.[0-9]\+" version/version.go)
SHELL=/bin/sh
ARCH=$(shell uname -m)
DIST_NAME=tmsu-$(ARCH)-$(VER)
DIST_DIR=$(DIST_NAME)
DIST_FILE=$(DIST_NAME).tgz

all: clean compile dist test

clean:
	@echo
	@echo "CLEANING"
	@echo
	go clean
	rm -Rf bin
	rm -Rf $(DIST_DIR)
	rm -f $(DIST_FILE)

compile:
	@echo
	@echo "COMPILING"
	@echo
	@mkdir -p bin
	go build -o bin/tmsu

test: unit-test integration-test

unit-test: compile
	@echo
	@echo "RUNNING UNIT TESTS"
	@echo
	go test ./...

integration-test: compile
	@echo
	@echo "RUNNING INTEGRATION TESTS"
	@echo
	@cd tests && ./runall

dist: compile
	@echo
	@echo "PACKAGING DISTRIBUTABLE"
	@echo
	@mkdir -p $(DIST_DIR)
	@mkdir -p $(DIST_DIR)/bin
	@mkdir -p $(DIST_DIR)/man
	@mkdir -p $(DIST_DIR)/misc/zsh
	@mkdir -p $(DIST_DIR)/misc/bash
	gcp -R bin -t $(DIST_DIR)
	gcp README.md -t $(DIST_DIR)
	gcp COPYING.md -t $(DIST_DIR)
	gcp misc/bin/* -t $(DIST_DIR)/bin/
	gzip -fc misc/man/tmsu.1 >$(DIST_DIR)/man/tmsu.1.gz
	gcp misc/zsh/_tmsu -t $(DIST_DIR)/misc/zsh/
	gcp misc/bash/tmsu -t $(DIST_DIR)/misc/bash/
	tar czf $(DIST_FILE) $(DIST_DIR)

install: 
	@echo
	@echo "INSTALLING"
	@echo
	mkdir -p $(INSTALL_DIR)
	mkdir -p $(MOUNT_INSTALL_DIR)
	mkdir -p $(MAN_INSTALL_DIR)
	mkdir -p $(ZSH_COMP_INSTALL_DIR)
	mkdir -p $(BASH_COMP_INSTALL_DIR)
	gcp bin/tmsu -t $(INSTALL_DIR)
	gcp misc/bin/mount.tmsu -t $(MOUNT_INSTALL_DIR)
	gcp misc/bin/tmsu-* -t $(INSTALL_DIR)
	gzip -fc misc/man/tmsu.1 >$(MAN_INSTALL_DIR)/tmsu.1.gz
	gcp misc/zsh/_tmsu -t $(ZSH_COMP_INSTALL_DIR)
	gcp misc/bash/tmsu -t $(BASH_COMP_INSTALL_DIR)

uninstall:
	@echo "UNINSTALLING"
	rm $(INSTALL_DIR)/tmsu
	rm $(MOUNT_INSTALL_DIR)/mount.tmsu
	rm $(INSTALL_DIR)/tmsu-*
	rm $(MAN_INSTALL_DIR)/tmsu.1.gz
	rm $(ZSH_COMP_INSTALL_DIR)/_tmsu
	rm $(BASH_COMP_INSTALL_DIR)/tmsu

.PHONY: all clean compile test unit-test integration-test dist install uninstall
