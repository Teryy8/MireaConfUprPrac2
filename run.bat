@echo off
cd /d "%~dp0"
if "%~1"=="--test" (
    go test ./...
) else (
    go run ./src %*
)
