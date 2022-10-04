.PHONY: run

run:
	go run ./main.go

.PHONY: build docker.run docker.up docker.down docker.rm

build:	## Build backend Docker image
	docker build . \
		-t website \
		--no-cache \

docker.run:
	docker run -d \
	-p 8001:8001 \
	--name website website

docker.up:
	docker container start website

docker.down:
	docker container stop website

docker.rm:
	docker rm -f website