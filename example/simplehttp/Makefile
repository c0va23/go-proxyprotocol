IMAGE_NAME := simplehttp

bin/:
	mkdir bin

bin/simplehttp: bin/ main.go
	CGO_ENABLED=0 go build -o bin/simplehttp main.go

build-image:
	docker build -t $(IMAGE_NAME) .
