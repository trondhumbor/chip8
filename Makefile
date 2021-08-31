build-dir:
	mkdir -p bin

build: build-dir
	go build -o bin/chip8 cmd/main/main.go

run: # make ARGS="-arg1 val1 -arg2 -arg3" run
	./bin/chip8 ${ARGS}

clean:
	rm -r bin
