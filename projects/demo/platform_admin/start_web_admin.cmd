@echo off
title Platform Admin - Web Admin (http://localhost:3000)
cd /d C:\data\developer\studio\games\aidev\projects\demo\platform_admin\dev-web-admin

echo Starting admin frontend on :3000 ...
pnpm dev

echo.
echo [Admin frontend exited] exit code = %errorlevel%
pause
