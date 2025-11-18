@echo off
REM Migration script for Windows

setlocal enabledelayedexpansion

REM Load environment variables from .env file
if exist .env (
    for /f "usebackq tokens=1,* delims==" %%a in (".env") do (
        set "line=%%a"
        if not "!line:~0,1!"=="#" (
            set "%%a=%%b"
        )
    )
)

REM Default values
if not defined DATABASE_URL set DATABASE_URL=postgres://postgres:postgres@localhost:5432/unchingspot?sslmode=disable

REM Check if migrate command exists
where migrate >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: golang-migrate CLI is not installed
    echo Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    exit /b 1
)

REM Parse command
set COMMAND=%1
if "%COMMAND%"=="" set COMMAND=up

if "%COMMAND%"=="up" (
    echo Running migrations up...
    migrate -path migrations -database "%DATABASE_URL%" up
    if %ERRORLEVEL% EQU 0 (
        echo Migrations completed successfully!
    )
) else if "%COMMAND%"=="down" (
    echo Rolling back migrations...
    migrate -path migrations -database "%DATABASE_URL%" down
    if %ERRORLEVEL% EQU 0 (
        echo Rollback completed successfully!
    )
) else if "%COMMAND%"=="force" (
    if "%2"=="" (
        echo Error: Please specify version to force
        echo Usage: scripts\migrate.bat force ^<version^>
        exit /b 1
    )
    echo Forcing migration version to %2...
    migrate -path migrations -database "%DATABASE_URL%" force %2
    if %ERRORLEVEL% EQU 0 (
        echo Force completed successfully!
    )
) else if "%COMMAND%"=="version" (
    echo Current migration version:
    migrate -path migrations -database "%DATABASE_URL%" version
) else if "%COMMAND%"=="create" (
    if "%2"=="" (
        echo Error: Please specify migration name
        echo Usage: scripts\migrate.bat create ^<migration_name^>
        exit /b 1
    )
    echo Creating new migration: %2
    migrate create -ext sql -dir migrations -seq %2
    if %ERRORLEVEL% EQU 0 (
        echo Migration files created successfully!
    )
) else (
    echo Usage: scripts\migrate.bat {up^|down^|force ^<version^>^|version^|create ^<name^>}
    exit /b 1
)

endlocal
