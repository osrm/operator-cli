build: cmd/operator-cli/main.go 
	go build -o operator-cli -ldflags "-X main.VERSION="$$(git describe --tags --abbrev=0) cmd/operator-cli/main.go 

test:
	./operator-cli deRegisterOperatorFromAVS --config-file template/config.json
	./operator-cli registerOperatorToAVS --config-file template/config.json
	./operator-cli deRegisterWatchtower --config-file template/config.json
	./operator-cli registerWatchtower --config-file template/config.json
