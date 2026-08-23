extends Node
## HTML-runtime PoC driver: proves the three headless-provable pieces of the
## HTML game runner:
##   1. zip package (.novahtml) round-trip via core ZIPPacker/ZIPReader
##   2. local HTTP file server (what a WebView plugin would talk to)
##      - single poll loop drives both server & client sockets (no nested coroutines)
##   3. nova:// bridge URL parsing (quit/save protocol)
## The WebView itself is platform-specific (gdcef / Android WebView / WKWebView)
## and is the only part not exercised here.

var failures: Array = []

func ok(cond: bool, label: String) -> void:
    print((("PASS - " if cond else "FAIL - ") + label))
    if not cond:
        failures.append(label)

func _ready() -> void:
    await get_tree().process_frame
    await _run()
    await get_tree().process_frame
    print("")
    if failures.is_empty():
        print("ALL HTML RUNTIME POC DONE")
        get_tree().quit(0)
    else:
        printerr("FAILURES: " + ", ".join(failures))
        get_tree().quit(1)

func _run() -> void:
    # ---------- 1. package a mini html game ----------
    var zip_path := "user://downloads/mini_html.novahtml"
    DirAccess.make_dir_recursive_absolute("user://downloads")
    var zp := ZIPPacker.new()
    var err := zp.open(zip_path)
    ok(err == OK, "ZIPPacker.open (%d)" % err)
    var index_html := "<!DOCTYPE html><html><body><script src=\"assets/app.js\"></script></body></html>"
    var meta_json := "{\"id\":\"mini_html\",\"runtime\":\"html\",\"entry\":\"index.html\"}"
    var app_js := "location.href='nova://quit?score=123'; // bridge example"
    zp.start_file("index.html")
    zp.write_file(index_html.to_utf8_buffer())
    zp.close_file()
    zp.start_file("meta.json")
    zp.write_file(meta_json.to_utf8_buffer())
    zp.close_file()
    zp.start_file("assets/app.js")
    zp.write_file(app_js.to_utf8_buffer())
    zp.close_file()
    zp.close()
    ok(FileAccess.file_exists(zip_path), "zip package written")

    # ---------- 2. extract to per-game dir ----------
    var game_dir := "user://games/mini_html/"
    DirAccess.make_dir_recursive_absolute(game_dir)
    var zr := ZIPReader.new()
    err = zr.open(zip_path)
    ok(err == OK, "ZIPReader.open (%d)" % err)
    var names: PackedStringArray = zr.get_files()
    ok(names.size() == 3, "zip lists 3 files (%d)" % names.size())
    for fn in names:
        var target := game_dir + fn
        if fn.contains("/"):
            DirAccess.make_dir_recursive_absolute(target.get_base_dir())
        var f := FileAccess.open(target, FileAccess.WRITE)
        f.store_buffer(zr.read_file(fn))
        f.close()
    zr.close()
    ok(FileAccess.file_exists(game_dir + "index.html"), "index.html extracted")
    ok(FileAccess.file_exists(game_dir + "assets/app.js"), "assets/app.js extracted")

    # ---------- 3. local http server ----------
    var server := TCPServer.new()
    var port := 0
    for try_port in [8917, 8918, 8919, 8920, 8921]:
        if server.listen(try_port) == OK:
            port = try_port
            break
    ok(port > 0, "http server listening (port=%d)" % port)
    if port == 0:
        return

    await _serve_request(server, port, "/games/mini_html/index.html",
        func(r: Dictionary) -> void:
            ok(r.code == 200, "GET index.html -> 200")
            ok(String(r.body).contains("<html>"), "index.html body served")
            ok(String(r.headers).to_lower().contains("text/html"), "index.html mime text/html")
    )
    await _serve_request(server, port, "/games/mini_html/assets/app.js",
        func(r: Dictionary) -> void:
            ok(r.code == 200, "GET app.js -> 200")
            ok(String(r.headers).to_lower().contains("text/javascript"), "app.js mime text/javascript")
    )
    await _serve_request(server, port, "/games/../shell_secret.txt",
        func(r: Dictionary) -> void:
            ok(r.code == 403, "path traversal blocked -> 403 (got %d)" % r.code)
    )
    await _serve_request(server, port, "/games/mini_html/missing.png",
        func(r: Dictionary) -> void:
            ok(r.code == 404, "missing file -> 404 (got %d)" % r.code)
    )

    # ---------- 4. nova:// bridge protocol ----------
    var b := parse_nova_url("nova://quit?score=1234&playtime=65.5&achievements=demo_a,demo_b")
    ok(b.get("cmd", "") == "quit", "bridge cmd=quit")
    ok(int(b.query.get("score", 0)) == 1234, "bridge score=1234")
    ok(absf(float(b.query.get("playtime", 0)) - 65.5) < 0.01, "bridge playtime=65.5")
    ok(String(b.query.get("achievements", "")).split(",").size() == 2, "bridge achievements split")
    var b2 := parse_nova_url("nova://save?key=best&value=999")
    ok(b2.cmd == "save" and b2.query.get("value") == "999", "bridge save command")
    ok(parse_nova_url("https://evil.com").is_empty(), "non-nova url rejected")
    server.stop()

# ---------------------------------------------------------------- http server
const MIME := {
    "html": "text/html; charset=utf-8",
    "js": "text/javascript",
    "mjs": "text/javascript",
    "css": "text/css",
    "json": "application/json",
    "png": "image/png",
    "jpg": "image/jpeg",
    "webp": "image/webp",
    "svg": "image/svg+xml",
    "wasm": "application/wasm",
    "ogg": "audio/ogg",
    "mp3": "audio/mpeg",
    "txt": "text/plain; charset=utf-8",
}

func _mime_for(path: String) -> String:
    return MIME.get(path.get_extension().to_lower(), "application/octet-stream")

func _respond(sock: StreamPeerTCP, req: String) -> void:
    var first := req.split("\r\n")[0]
    var parts: PackedStringArray = first.split(" ")
    var code := 400
    var local := ""
    if parts.size() >= 2:
        var path := parts[1]
        code = 200
        if path.contains("..") or path.contains("\\") or path.contains("%00"):
            code = 403
        else:
            local = "user://" + path.lstrip("/")
            if not FileAccess.file_exists(local):
                code = 404
    var body := PackedByteArray()
    if code == 200:
        var f := FileAccess.open(local, FileAccess.READ)
        body = f.get_buffer(f.get_length())
        f.close()
    var reason: String = {200: "OK", 403: "Forbidden", 404: "Not Found"}.get(code, "Bad Request")
    var header := "HTTP/1.1 %d %s\r\n" % [code, reason]
    header += "Content-Type: %s\r\n" % _mime_for(local if local != "" else "a.png")
    header += "Content-Length: %d\r\n" % body.size()
    header += "Access-Control-Allow-Origin: *\r\n"
    header += "Connection: close\r\n\r\n"
    sock.put_data(header.to_utf8_buffer())
    if body.size() > 0:
        sock.put_data(body)

func _body_complete(text: String) -> bool:
    var idx := text.find("\r\n\r\n")
    if idx < 0:
        return false
    var head := text.substr(0, idx)
    var body := text.substr(idx + 4)
    var cl := 0
    for line in head.split("\r\n"):
        var l := line.to_lower()
        if l.begins_with("content-length:"):
            cl = int(l.substr(l.find(":") + 1).strip_edges())
    return body.length() >= cl

# one GET request through server+client, single manual poll loop
func _serve_request(server: TCPServer, port: int, get_path: String, check: Callable) -> void:
    var peer := StreamPeerTCP.new()
    peer.connect_to_host("127.0.0.1", port)
    var sent := false
    var resp := PackedByteArray()
    var sock: StreamPeerTCP = null
    var responded := false
    for i in 4000:
        # --- server side ---
        if sock == null and server.is_connection_available():
            sock = server.take_connection()
        if sock != null and not responded:
            var n := sock.get_available_bytes()
            if n > 0:
                var req := sock.get_utf8_string(n)
                if (req).contains("\r\n\r\n"):
                    _respond(sock, req)
                    responded = true
        # --- client side ---
        peer.poll()
        if peer.get_status() == StreamPeerTCP.STATUS_CONNECTED and not sent:
            var head := "GET %s HTTP/1.1\r\nHost: 127.0.0.1\r\nConnection: close\r\n\r\n" % get_path
            peer.put_data(head.to_utf8_buffer())
            sent = true
        var av := peer.get_available_bytes()
        if av > 0:
            resp.append_array(peer.get_partial_data(av)[1])
        if resp.size() > 0 and _body_complete(resp.get_string_from_utf8()):
            break
        await get_tree().process_frame
    var text := resp.get_string_from_utf8()
    var code := 0
    var headers := ""
    var body := ""
    if text != "":
        var hs := text.split("\r\n\r\n", true, 1)
        headers = hs[0]
        if hs.size() > 1:
            body = hs[1]
        var words: PackedStringArray = headers.split("\r\n")[0].split(" ")
        if words.size() >= 2:
            code = int(words[1])
    check.call({code = code, headers = headers, body = body})

# ---------------------------------------------------------------- nova:// bridge
func parse_nova_url(url: String) -> Dictionary:
    if not url.begins_with("nova://"):
        return {}
    var rest := url.substr("nova://".length())
    var cmd := rest
    var query := {}
    var q := rest.find("?")
    if q >= 0:
        cmd = rest.substr(0, q)
        for pair in rest.substr(q + 1).split("&"):
            var kv: PackedStringArray = pair.split("=", true, 1)
            if kv.size() == 2:
                query[kv[0]] = kv[1].uri_decode()
    return {"cmd": cmd, "query": query}
