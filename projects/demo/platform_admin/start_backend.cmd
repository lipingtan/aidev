@echo off
title Platform Admin - Backend (http://localhost:8000)
cd /d "%~dp0backend"

echo Starting Go backend on :8000 ...
go run main.go server -c config/settings.yml

echo.
echo [Backend exited] exit code = %errorlevel%
pause
