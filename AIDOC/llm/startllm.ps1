& "C:\data\llama.cpp\llama-server.exe" `
    --models-preset "C:\data\run\models.ini" `
    --host 0.0.0.0 `
    --port 8080 `
    --jinja `
	--flash-attn on `
    --ui-mcp-proxy `
	--cache-reuse 512  `
	--no-mmproj-offload  `
    --load-mode mmap  `
	-np 1 
	