make: example example-multi-arg
example:
	go build -o example-gamma-client ./cmd/example-gamma-client/main.go
example-multi-arg:
	go build -o example-multi-arg-tool cmd/example-multi-arg-tool/main.go
clean:
	test -f ./example-gamma-client && rm ./example-gamma-client || true
	test -f ./example-multi-arg-tool && rm ./example-multi-arg-tool || true
