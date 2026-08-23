class_name CoreManager extends RefCounted
## CoreManager 桩（Q3=A）：模拟器核心安装状态查询与下载接口
##
## Mock 实现恒返回"未安装"并 push_warning 记录日志，不阻塞任何流程（RG-6）。
## 真实实现待 mame-godot-plugin v2 验证后填（runtime-design §3 临时保留旧口径待调整）。

## 查询指定版本核心是否已安装（Mock 恒返回未安装）
func is_installed(version: String) -> bool:
	push_warning("[CoreManager] Mock 查询：is_installed(%s) → 未安装" % version)
	return false

## 主动检查安装状态（Mock 恒返回未安装）
func check_installed(version: String) -> bool:
	push_warning("[CoreManager] Mock 检查：check_installed(%s) → 未安装" % version)
	return false

## 下载指定版本核心并按 sha256 校验（Mock 仅记录日志，不做真实下载）
func download(version: String, sha256: String) -> void:
	push_warning("[CoreManager] Mock 下载：version=%s sha256=%s（未执行真实下载）" % [version, sha256])

## 移除指定版本核心（Mock 仅记录日志，不做真实删除）
func remove(version: String) -> void:
	push_warning("[CoreManager] Mock 移除：version=%s（未执行真实删除）" % version)
