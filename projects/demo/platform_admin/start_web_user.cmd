@echo off
title Platform Admin - Web User (http://localhost:5174)
cd /d C:\data\developer\studio\games\aidev\projects\demo\platform_admin\dev-web-user

echo Starting user frontend on :5174 ...
pnpm dev

echo.
echo [User frontend exited] exit code = %errorlevel%
pause
