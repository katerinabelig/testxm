# Build the Go application
build:
	@go build -o bin/gocompany

# Run the Go application locally
run: build
	@./bin/gocompany

# # Run tests
# test:
# 	@go test -v ./...

# # Build the Docker image
# docker-build:
# 	@docker build -t gocompany:1.0 .

# # Run the PostgreSQL container
# db-run:
# 	@docker run --name postgres-container -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=gocompany -e POSTGRES_DB=postgres -p 5432:5432 -d postgres

# # Run the Go application container
# docker-run: docker-build db-run
# 	@docker run -p 3000:3000 -d gocompany:1.0

# # Clean up Docker containers and images
# docker-clean:
# 	@docker stop $$(docker ps -q) || true
# 	@docker rm $$(docker ps -a -q) || true

# # All-in-one command to build, create Docker image, and run the containers
# all: docker-clean docker-run