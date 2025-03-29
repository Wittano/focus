OUTPUT_DIR=./build

gen-templ:
	templ generate

build:
	go build -o $(OUTPUT_DIR)/app cmd/app/main.go

server: gen-templ
	go run cmd/server/main.go

cli:
	go run cmd/cli/main.go

clean:
	if [ -d ./build ]; then rm -r $(OUTPUT_DIR); fi
