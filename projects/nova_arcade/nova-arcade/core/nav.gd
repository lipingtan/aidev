extends Node
## 页面栈 + Tab 导航（push/pop/pop_to_root/switch_tab/current，返回键拦截）
##
## Autoload 名 Nav（注册顺序第 5），不声明 class_name（godot-engine §六）。
## 跳转只走本单例、通知只走 EventBus；页面间禁止互相引用。
## - push：实例化场景入栈，220ms 滑入转场，调 on_enter(data)
## - pop / pop_to_root：出栈，调 on_exit，恢复上层 on_resume
## - switch_tab：清栈回该 Tab 根页，发 EventBus.tab_changed(idx)
## - handle_back：模态 Dialog → 当前页 on_back() → 默认 pop
##
## 依赖：
## - EventBus: 发 tab_changed
## - Page: 页面基类（on_enter/on_exit/on_resume/on_back）

## 转场时长（秒），对应 220ms 滑入/滑出
const TRANSITION_MS: float = 0.22
## 滑入起点 X 偏移（视口宽 720，从右侧滑入）
const SLIDE_OFFSET: float = 720.0

# === Tab 根页映射（实例变量：_ready 初始化，便于测试用占位场景覆盖）===
var _tab_roots: Dictionary = {}
# === 页面栈（_pages 与 _paths 平行，top 在末尾）===
var _pages: Array[Page] = []
var _paths: Array[String] = []
# === PageStack 容器节点（占位场景缺失时为 null）===
var _page_stack: Control = null
var _current_tab: int = -1

## A-3：OverlayLayer 路径缓存（T10 执行时实查确认：/root/Main/App/OverlayLayer）
var _overlay_layer: Control = null

func _ready() -> void:
	_init_tab_roots()
	locate_page_stack()
	call_deferred("_cache_overlay_layer")

## 延迟缓存 OverlayLayer，确保主场景树完整后再查
func _cache_overlay_layer() -> void:
	_overlay_layer = get_node_or_null("/root/Main/App/OverlayLayer") as Control

## 初始化 Tab 根页映射（_ready 时；测试可覆盖 _tab_roots 后重查）
func _init_tab_roots() -> void:
	_tab_roots = {
		0: "res://shell/pages/home.tscn",
		1: "res://shell/pages/category.tscn",
		2: "res://shell/pages/search.tscn",
		3: "res://shell/pages/library.tscn",
	}

## 重新定位 PageStack 容器；占位/测试场景在 Nav._ready 之后才挂载，需手动再查一次。
func locate_page_stack() -> bool:
	var found := get_tree().root.find_child("PageStack", true, false) as Control
	if found == null:
		push_warning("Nav: 未找到 PageStack 节点（占位场景无此节点，导航暂不可用）")
		_page_stack = null
		return false
	_page_stack = found
	return true

## 压入页面：实例化 page_path，220ms 滑入，调 on_enter(data)
func push(page_path: String, data: Dictionary = {}) -> void:
	if _page_stack == null:
		push_warning("Nav.push: PageStack 未就绪，跳过")
		return
	var res := load(page_path) as PackedScene
	if res == null:
		push_error("Nav.push: 场景加载失败 %s" % page_path)
		return
	var page := res.instantiate() as Page
	if page == null:
		push_error("Nav.push: 非 Page 实例 %s" % page_path)
		return
	_pages.append(page)
	_paths.append(page_path)
	_page_stack.add_child(page)
	page.on_enter(data)
	_animate_in(page)

## 弹出顶层页（根页不可弹）；调 on_exit，恢复上层 on_resume
func pop() -> void:
	if _pages.size() <= 1:
		push_warning("Nav.pop: 已在根页")
		return
	var top: Page = _pages.pop_back()
	_paths.pop_back()
	if not _pages.is_empty():
		_pages.back().on_resume()
	_animate_out(top)

## 弹回到当前 Tab 根页（保留栈底）
func pop_to_root() -> void:
	while _pages.size() > 1:
		var top: Page = _pages.pop_back()
		_paths.pop_back()
		top.on_exit()
		top.queue_free()
	if not _pages.is_empty():
		_pages.back().on_resume()

## 切换 Tab：清栈回该 Tab 根页，发 EventBus.tab_changed(idx)
func switch_tab(idx: int) -> void:
	if not _tab_roots.has(idx):
		push_warning("Nav.switch_tab: 未知 Tab %d" % idx)
		return
	for p in _pages:
		(p as Page).on_exit()
		(p as Node).queue_free()
	_pages.clear()
	_paths.clear()
	_current_tab = idx
	_load_root(idx)
	EventBus.tab_changed.emit(idx)

## 返回键拦截：模态 Dialog → 当前页 on_back() → 默认 pop。return true 表示已消费。
func handle_back() -> bool:
	if _modal_dialog_open():
		return true
	var top: Page = null
	if not _pages.is_empty():
		top = _pages.back()
	if top != null and top.on_back():
		return true
	if _pages.size() <= 1:
		return false
	pop()
	return true

## 当前页路径；栈空返回 ""
func current() -> String:
	if _paths.is_empty():
		return ""
	return _paths.back()

# === 私有 ===

## 加载指定 Tab 的根页入栈（switch_tab 用）
func _load_root(idx: int) -> void:
	var path := _tab_roots[idx] as String
	if _page_stack == null:
		push_warning("Nav._load_root: PageStack 未就绪")
		return
	var res := load(path) as PackedScene
	if res == null:
		push_error("Nav._load_root: 根页加载失败 %s" % path)
		return
	var page := res.instantiate() as Page
	if page == null:
		push_error("Nav._load_root: 非 Page 实例 %s" % path)
		return
	_pages.append(page)
	_paths.append(path)
	_page_stack.add_child(page)
	page.on_enter({})

## 滑入动画（220ms，从右侧 SLIDE_OFFSET 到 0）
func _animate_in(page: Page) -> void:
	var tw := create_tween()
	tw.tween_property(page, "position:x", 0.0, TRANSITION_MS).from(SLIDE_OFFSET)

## 滑出动画（220ms，滑到右侧后释放）
func _animate_out(page: Page) -> void:
	page.on_exit()
	var tw := create_tween()
	tw.tween_property(page, "position:x", SLIDE_OFFSET, TRANSITION_MS).from(page.position.x)
	tw.tween_callback(page.queue_free)

## 模态 Dialog 查询：OverlayLayer 下任一子节点 visible=true 时返回 true（A-3：遍历方式，新增弹窗无需改此函数）
func _modal_dialog_open() -> bool:
	if _overlay_layer == null:
		return false
	for child in _overlay_layer.get_children():
		if child.visible:
			return true
	return false
