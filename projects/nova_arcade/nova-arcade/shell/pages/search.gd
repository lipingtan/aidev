extends Page
## 搜索页骨架（CR-1）：标题 + EmptyState 占位；on_enter 打印参数供冒烟验证

func on_enter(data: Dictionary) -> void:
	print("[search] on_enter ", str(data))
