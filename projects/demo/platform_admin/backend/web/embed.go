package web

import "embed"

// Files 嵌入前端构建产物
// 构建前需先执行：cd ../frontend && npm run build:prod
// 然后将 dist/ 内容复制到此目录的 dist/ 子目录
//
//go:embed dist
var Files embed.FS
