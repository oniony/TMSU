build:
	mkdir -p obj
	mkdir -p bin
	8g -o obj/commands.8 src/commands/help.go
	8g -o obj/main.8 src/main/main.go
	8l -o bin/tmsu obj/main.8
