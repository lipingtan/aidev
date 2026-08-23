class_name DummyPageB extends Page
## T5 回归测试用占位页 B（记录四接口调用，保留复用）

## on_enter 收到的参数
var enter_data: Dictionary = {}
## on_exit 调用次数
var exit_count: int = 0
## on_resume 调用次数
var resume_count: int = 0

func on_enter(data: Dictionary) -> void:
	enter_data = data

func on_exit() -> void:
	exit_count += 1

func on_resume() -> void:
	resume_count += 1
