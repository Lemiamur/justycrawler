# JustyCrawler - Web Crawler in Go

## Project Overview

JustyCrawler is a web crawler written in Go that crawls websites and stores the discovered links in MongoDB. It uses Redis to track visited URLs to avoid duplicate processing. The crawler is designed with a modular architecture, separating concerns into distinct packages for configuration, fetching, parsing, storage, and state management.

Key features:
- Concurrent crawling with configurable worker count
- Depth-limited crawling with optional same-host restriction
- MongoDB storage for crawled data
- Redis state tracking to prevent duplicate processing
- Graceful shutdown handling
- Configurable via YAML, environment variables, or command-line flags

## Project Structure

```
.
├── cmd/
│   └── main.go              # Application entry point
├── configs/
│   └── config.yaml          # Default configuration file
├── internal/
│   ├── app/
│   │   └── crawler/         # Core crawler logic
│   ├── config/              # Configuration management
│   ├── domain/              # Domain entities
│   ├── fetcher/             # HTTP fetching implementation
│   ├── parser/              # HTML parsing implementation
│   ├── state/               # Redis state management
│   └── storage/             # MongoDB storage implementation
├── mocks/                   # Generated mocks for testing
├── go.mod                   # Go module dependencies
├── go.sum                   # Go module checksums
├── makefile                 # Build and run commands
├── docker-compose.yml       # Docker services configuration
└── .golangci.yml            # Linting configuration
```

## Technologies Used

- **Go 1.23**: Programming language
- **MongoDB**: Storage for crawled data
- **Redis**: State tracking for visited URLs
- **goquery**: HTML parsing library
- **Viper**: Configuration management
- **pflag**: Command-line flag parsing
- **golangci-lint**: Code linting

## Domain Model

The crawler works with a simple data model:

```go
type CrawledData struct {
    URL        string   `bson:"url"`
    Depth      int      `bson:"depth"`
    FoundOn    string   `bson:"found_on"` // URL where this page was found
    FoundLinks []string `bson:"found_links"`
}
```

## Core Components

### 1. Configuration (`internal/config`)
- Uses Viper for configuration management
- Supports configuration via:
  - YAML config file (`configs/config.yaml`)
  - Environment variables
  - Command-line flags
- Priority order: flags > environment variables > config file

### 2. Crawler (`internal/app/crawler`)
- Main crawling logic with concurrent workers
- Implements depth-limited crawling
- Optional same-host restriction
- Task queue management with channels
- Graceful shutdown handling

### 3. Fetcher (`internal/fetcher`)
- HTTP fetching implementation using `net/http`
- Configurable timeouts
- User-Agent header for requests

### 4. Parser (`internal/parser`)
- HTML parsing using `goquery`
- Extracts all valid HTTP/HTTPS links from pages
- Resolves relative URLs to absolute URLs
- Deduplicates found links

### 5. Storage (`internal/storage`)
- MongoDB implementation for storing crawled data
- Upsert operations to avoid duplicates
- Connection management with proper cleanup

### 6. State (`internal/state`)
- Redis implementation for tracking visited URLs
- Uses Redis Sets for atomic operations
- Connection management with proper cleanup

## Building and Running

### Prerequisites
- Go 1.23+
- Docker and Docker Compose (for dependencies)

### Running with Docker (Recommended)
First, start the dependencies:
```bash
docker-compose up -d mongodb redis
```

Then run the application:
```bash
make run
```

### Local Development
1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Start dependencies:
   ```bash
   docker-compose up -d mongodb redis
   ```

3. Run the application:
   ```bash
   go run ./cmd/main.go --start_url https://example.com
   ```

### Building
```bash
make build
# or
go build -o ./bin/crawler ./cmd/main.go
```

### Available Make Commands
- `make run`: Run the application with force recrawl
- `make build`: Build the application
- `make test`: Run tests (if any)
- `make lint`: Run linter
- `make mocks`: Generate mocks
- `make tidy`: Tidy Go modules

## Configuration

The application can be configured through multiple methods with the following priority:

1. Command-line flags
2. Environment variables
3. YAML configuration file (`configs/config.yaml`)

### Key Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `start_url` | Starting URL for crawling | Required |
| `same_host` | Restrict crawling to same host | true |
| `max_depth` | Maximum crawl depth | 2 |
| `worker_count` | Number of concurrent workers | 10 |
| `force_recrawl` | Clear state before crawling | false |
| `http.timeout` | HTTP request timeout | 30s |
| `mongo.uri` | MongoDB connection URI | mongodb://localhost:27017 |
| `mongo.database` | MongoDB database name | crawler_db |
| `mongo.collection` | MongoDB collection name | links |
| `redis.addr` | Redis address | localhost:6379 |
| `redis.password` | Redis password | "" |
| `redis.db` | Redis database number | 0 |
| `redis.set_key` | Redis set key for visited URLs | crawler:visited_urls |
| `log.level` | Logging level | info |

Environment variables use underscores instead of dots (e.g., `MONGO_URI` instead of `mongo.uri`).

## Development Conventions

1. **Code Structure**: Follows clean architecture principles with clear separation of concerns
2. **Interfaces**: All external dependencies are defined as interfaces for testability
3. **Error Handling**: Proper error wrapping with context
4. **Logging**: Uses structured logging with `log/slog`
5. **Testing**: Uses mocks for unit testing (generated with mockery)
6. **Linting**: Strict linting with golangci-lint
7. **Dependency Management**: Go modules for dependency management

## Testing

Run tests with:
```bash
make test
# or
go test -v ./...
```

Generate mocks with:
```bash
make mocks
# or
go generate ./...
```

## Linting

Run linter with:
```bash
make lint
# or
golangci-lint run
```

## Docker Services

The `docker-compose.yml` file defines the following services:
- MongoDB: For storing crawled data
- Redis: For tracking visited URLs

To start services:
```bash
docker-compose up -d
```

## Typical Usage

1. Start dependencies:
   ```bash
   docker-compose up -d mongodb redis
   ```

2. Run the crawler:
   ```bash
   go run ./cmd/main.go --start_url https://example.com --max_depth 2
   ```

3. To force a recrawl (clear previous state):
   ```bash
   go run ./cmd/main.go --start_url https://example.com --force_recrawl
   ```

## Commenting Principles

When adding or updating comments in the codebase, follow these principles:

1. Write brief and informative comments that explain not what the code does, but why it does it
2. Use simple and clear language, avoiding overly technical jargon
3. Update comments as the code changes
4. Be consistent in commenting style throughout the entire codebase