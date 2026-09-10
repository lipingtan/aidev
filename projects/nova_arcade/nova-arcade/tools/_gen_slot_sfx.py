# 一次性生成 CR-8 水果机音效（纯 stdlib 合成 WAV，无需三方库）：
#   win_big.wav  —— 中奖震撼音：上行琶音 + 双鼓点（~1.1s）
#   orb_run.wav  —— 散灯跑动低频轰鸣（~1.0s，可循环感）
# 复现：python tools/_gen_slot_sfx.py
import math
import random
import struct
import wave

SR = 22050  # 采样率


def write_wav(path: str, samples: list[float]) -> None:
    frames = b"".join(struct.pack("<h", max(-32767, min(32767, int(s * 32767)))) for s in samples)
    with wave.open(path, "wb") as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(SR)
        w.writeframes(frames)
    print("wrote", path, f"{len(samples) / SR:.2f}s")


def tone(freq: float, dur: float, vol: float = 0.6, decay: float = 6.0) -> list[float]:
    n = int(SR * dur)
    out = []
    for i in range(n):
        t = i / SR
        env = math.exp(-decay * t)
        # 基频 + 泛音，听感更亮
        v = math.sin(2 * math.pi * freq * t) + 0.35 * math.sin(2 * math.pi * freq * 2 * t)
        out.append(vol * env * v / 1.35)
    return out


def drum(dur: float = 0.18, vol: float = 0.9, punch: float = 90.0) -> list[float]:
    n = int(SR * dur)
    out = []
    random.seed(7)
    for i in range(n):
        t = i / SR
        f = punch * math.exp(-18 * t) + 45.0
        env = math.exp(-14 * t)
        noise = (random.random() * 2 - 1) * 0.25 * math.exp(-40 * t)
        out.append(vol * env * math.sin(2 * math.pi * f * t) + noise)
    return out


def mix(dst: list[float], src: list[float], at: float) -> None:
    o = int(at * SR)
    for i, s in enumerate(src):
        j = o + i
        if j < len(dst):
            dst[j] += s


# --- win_big：C5-E5-G5-C6 琶音 + 两记鼓点 ---
total = int(SR * 1.1)
buf = [0.0] * total
for k, f in enumerate([523.25, 659.25, 783.99, 1046.5]):
    mix(buf, tone(f, 0.5, 0.5, 5.0), 0.10 * k)
mix(buf, drum(), 0.0)
mix(buf, drum(), 0.42)
# 轻限幅
buf = [max(-0.95, min(0.95, s)) for s in buf]
write_wav("games/slot_machine/assets/sounds/win_big.wav", buf)

# --- orb_run：低频锯齿轰鸣 + 颤动（散灯跑动循环感）---
total = int(SR * 1.0)
buf = [0.0] * total
for i in range(total):
    t = i / SR
    f = 55.0 + 18.0 * math.sin(2 * math.pi * 7.0 * t)   # 55Hz 低频 + 7Hz 颤动
    saw = 2.0 * ((f * t) % 1.0) - 1.0                    # 锯齿波
    trem = 0.6 + 0.4 * math.sin(2 * math.pi * 13.0 * t)  # 音量颤抖
    env = min(1.0, t * 8.0) * min(1.0, (1.0 - t) * 8.0 + 0.15)
    buf[i] = 0.5 * env * trem * saw
write_wav("games/slot_machine/assets/sounds/orb_run.wav", buf)

# --- firecracker：礼炮爆响（白噪爆裂 + 低频冲击 + 混响尾巴，~0.6s）---
total = int(SR * 0.6)
buf = [0.0] * total
random.seed(42)
for i in range(total):
    t = i / SR
    # 1) 白噪爆裂（前 60ms 快速衰减）
    crack = (random.random() * 2 - 1) * math.exp(-55 * t)
    # 2) 低频冲击（punch 下扫 120→40Hz）
    f = 40.0 + 80.0 * math.exp(-30 * t)
    boom = 0.9 * math.exp(-12 * t) * math.sin(2 * math.pi * f * t)
    # 3) 混响尾巴（衰减噪声，慢包络）
    tail = (random.random() * 2 - 1) * 0.12 * math.exp(-6 * t) * (0.5 + 0.5 * math.sin(2 * math.pi * 31 * t))
    buf[i] = crack + boom + tail
buf = [max(-0.95, min(0.95, s * 0.9)) for s in buf]
write_wav("games/slot_machine/assets/sounds/firecracker.wav", buf)
