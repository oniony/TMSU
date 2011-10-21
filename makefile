tmsu:
	mkdir -p obj
	8g -o obj/tmsu.8 src/main.go
	mkdir -p bin
	8l -o bin/tmsu obj/tmsu.8

clean:
	rm -f bin obj
