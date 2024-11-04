build:
	go build -o operator-cli -ldflags "-X main.VERSION="$$(git describe --tags --abbrev=0) cmd/operator-cli/main.go 

test: build
	./operator-cli registerWatchtower --config-file template/config.json
	./operator-cli deRegisterWatchtower --config-file template/config.json
	# ./operator-cli deRegisterOperatorFromAVS --config-file template/holesky.json
	# ./operator-cli registerOperatorToAVS --config-file template/holesky.json
