.PHONY: build install clean test run

# Binary name
BINARY_NAME=apeople

# Build the binary
build:
	go build -o $(BINARY_NAME) .

# Install to GOPATH/bin
install:
	go install .

# Install to /usr/local/bin (requires sudo)
install-system: build
	sudo cp $(BINARY_NAME) /usr/local/bin/

# Install to ~/bin
install-user: build
	mkdir -p ~/bin
	cp $(BINARY_NAME) ~/bin/

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	go clean

# Run tests
test:
	go test ./...

# Run the application
run: build
	./$(BINARY_NAME)

# Setup configuration
setup-config:
	./setup-config.sh

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 .