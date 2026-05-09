@echo off
cd /d "%~dp0"
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080" ^| findstr "LISTENING"') do (
  echo Stopping process on port 8080 (PID %%a^)...
  taskkill /PID %%a /F >nul 2>nul
)
go run ./src/backend
