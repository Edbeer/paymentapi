build:
	@go build -o ./bin/api

run: build
	@./bin/api

docker:
	docker run --name payment-postgres \
	-e POSTGRES_PASSWORD=postgres \
	-e POSTGRES_USER=postgres \
	-e POSTGRES_DB=paymentdb \
	-p 5432:5432 -d postgres

docker-start:
	docker start payment-postgres

docker-exec: docker-start
	docker exec -it payment-postgres psql -U postgres paymentdb

docker-stop:
	docker stop payment-postgres