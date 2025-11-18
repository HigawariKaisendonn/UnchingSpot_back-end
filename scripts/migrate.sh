#!/bin/bash

# Migration script for Unix-like systems (Linux, macOS)

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default values
DATABASE_URL=${DATABASE_URL:-"postgres://postgres:postgres@localhost:5432/unchingspot?sslmode=disable"}
MIGRATIONS_PATH=${MIGRATIONS_PATH:-"file://migrations"}

# Check if migrate command exists
if ! command -v migrate &> /dev/null; then
    echo "Error: golang-migrate CLI is not installed"
    echo "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Parse command
COMMAND=${1:-"up"}

case $COMMAND in
    up)
        echo "Running migrations up..."
        migrate -path migrations -database "$DATABASE_URL" up
        echo "Migrations completed successfully!"
        ;;
    down)
        echo "Rolling back migrations..."
        migrate -path migrations -database "$DATABASE_URL" down
        echo "Rollback completed successfully!"
        ;;
    force)
        if [ -z "$2" ]; then
            echo "Error: Please specify version to force"
            echo "Usage: ./scripts/migrate.sh force <version>"
            exit 1
        fi
        echo "Forcing migration version to $2..."
        migrate -path migrations -database "$DATABASE_URL" force $2
        echo "Force completed successfully!"
        ;;
    version)
        echo "Current migration version:"
        migrate -path migrations -database "$DATABASE_URL" version
        ;;
    create)
        if [ -z "$2" ]; then
            echo "Error: Please specify migration name"
            echo "Usage: ./scripts/migrate.sh create <migration_name>"
            exit 1
        fi
        echo "Creating new migration: $2"
        migrate create -ext sql -dir migrations -seq $2
        echo "Migration files created successfully!"
        ;;
    *)
        echo "Usage: ./scripts/migrate.sh {up|down|force <version>|version|create <name>}"
        exit 1
        ;;
esac
