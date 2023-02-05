build:
	go build -o cmd/server/server cmd/server/main.go
	# build the wasm code :)
	GOOS=js GOARCH=wasm go build -o assets/json.wasm cmd/wasm/main.go

clean:
	rm -rf assets/json.wasm
	rm -rf cmd/server/server

