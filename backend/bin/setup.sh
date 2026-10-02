cp .env.example .env
go mod tidy
go test ./...              # util + game + store tests
go run ./cmd/cli seed      # inserts the starter Belarusian cities
go run ./cmd/server        # :8080
