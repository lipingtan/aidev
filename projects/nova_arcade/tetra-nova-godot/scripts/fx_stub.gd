extends Node
## No-op FX stub for headless logic tests.

var beat := 0.0

func shake(_m: float) -> void: pass
func stop(_t: float) -> void: pass
func flash(_col: Color, _a: float) -> void: pass
func slowmo(_t: float, _s: float) -> void: pass
func pop_center(_txt: String, _col: Color, _size: int, _dx: int, _dy: int) -> void: pass
func pop_board(_gx: float, _gy: float, _txt: String, _col: Color, _size: int) -> void: pass
func burst_cell(_x: int, _y: int, _col: Color, _n: int, _pow: float) -> void: pass
func burst_board(_gx: float, _gy: float, _col: Color, _n: int, _pow: float) -> void: pass
func ring_cell(_x: int, _y: int, _col: Color, _r: float, _v: float, _w: float, _life: float) -> void: pass
func ring_board(_gx: float, _gy: float, _col: Color, _r: float, _v: float, _w: float, _life: float) -> void: pass
func ring_center(_col: Color, _r: float, _v: float, _w: float, _life: float) -> void: pass
func bolt(_tx: int, _ty: int) -> void: pass
