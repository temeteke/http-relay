# HTTP Relay

This project is an HTTP relay server that proxies requests to target URLs and modifies the response body to rewrite URLs.

## Prerequisites

- Go 1.23.4 or later

## Getting Started

### Clone the repository

```sh
git clone https://github.com/temeteke/http-relay.git
cd http-relay
```

### Build the project

```sh
go build -o http-relay
```

### Run the server

```sh
./http-relay
```

The server will start listening on port 8080.

## Usage

To use the HTTP relay, make a request to the server with the target URL as the path. For example:

```sh
curl http://localhost:8080/http://example.com
```

The server will proxy the request to `http://example.com` and return the response, modifying any URLs in the response body to point back to the relay server.

## Development

### Dev Container

This project includes a dev container configuration for Visual Studio Code. To use it, open the project in Visual Studio Code and select "Reopen in Container" when prompted.

### Dependencies

This project uses the following dependencies:

- [github.com/gorilla/mux](https://github.com/gorilla/mux) v1.8.1

### Project Structure

- `main.go`: The main entry point of the application.
- `.devcontainer/devcontainer.json`: Configuration for the development container.