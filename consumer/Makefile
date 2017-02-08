PHONY: compile build push
compile:
	go build main.go 

build:
	docker build -t consumer:latest .

push:
	docker tag consumer knabben/consumer
	docker push knabben/consumer
