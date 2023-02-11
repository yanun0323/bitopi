.PHONY: run

run:
	go run ./main.go

.PHONY: build docker.run docker.up docker.down docker.rm

build:	## Build backend Docker image
	docker build . \
		-t bitopi \
		--no-cache \

docker.run:
	docker run -d \
	-p 8001:8001 \
	--link mysql:mysql \
	--name bitopi bitopi

docker.up:
	docker container start bitopi

docker.down:
	docker container stop bitopi

docker.rm:
	docker rm -f bitopi

.PHONY: mysql.run
mysql.run:
	docker run -d -p 3306:3306 \
		-e MYSQL_USER=${USER} -e MYSQL_PASSWORD=${PASSWARD} \
    	-e MYSQL_DATABASE=${DATABASE} \
    	--name mysql mysql/mysql-server:8.0