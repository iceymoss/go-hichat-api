/**
 * 通话铃声 —— Web Audio API 合成，无需音频文件、离线可用。
 *
 * - playRingback()：主叫回铃音（呼出等待对方接听）
 * - playRingtone()：被叫来电铃声（+ 移动端振动）
 * - stopRing()：停止当前铃声
 *
 * 注意自动播放策略：AudioContext 需在用户手势后才能出声。主叫点了「呼叫」即有手势；
 * 被叫在已使用 App 的前提下大多已解锁；suspended 时尝试 resume，失败则静默。
 */

type WindowWithWebkitAudio = Window & { webkitAudioContext?: typeof AudioContext };

let ctx: AudioContext | null = null;
let timer: ReturnType<typeof setInterval> | null = null;
let vibrateTimer: ReturnType<typeof setInterval> | null = null;

function getCtx(): AudioContext | null {
  if (typeof window === 'undefined') return null;
  if (!ctx) {
    const AC = window.AudioContext || (window as WindowWithWebkitAudio).webkitAudioContext;
    if (!AC) return null;
    ctx = new AC();
  }
  if (ctx.state === 'suspended') ctx.resume().catch(() => { /* 等待手势 */ });
  return ctx;
}

/** 播放一段（可多频）正弦音，带淡入淡出避免爆音 */
function beep(c: AudioContext, freqs: number[], start: number, dur: number, vol: number) {
  const gain = c.createGain();
  gain.connect(c.destination);
  gain.gain.setValueAtTime(0.0001, start);
  gain.gain.exponentialRampToValueAtTime(vol, start + 0.03);
  gain.gain.setValueAtTime(vol, start + dur - 0.05);
  gain.gain.exponentialRampToValueAtTime(0.0001, start + dur);
  freqs.forEach(f => {
    const osc = c.createOscillator();
    osc.type = 'sine';
    osc.frequency.value = f;
    osc.connect(gain);
    osc.start(start);
    osc.stop(start + dur + 0.02);
  });
}

function startLoop(cycle: (c: AudioContext) => void, periodMs: number) {
  stopRing();
  const c = getCtx();
  if (!c) return;
  cycle(c);
  timer = setInterval(() => {
    const cc = getCtx();
    if (cc) cycle(cc);
  }, periodMs);
}

/** 主叫回铃音：440+480Hz，响 1s 停 3s（经典 ringback，音量偏低） */
export function playRingback() {
  startLoop((c) => {
    beep(c, [440, 480], c.currentTime + 0.02, 1.0, 0.1);
  }, 4000);
}

/** 被叫来电铃声：三连音循环 + 移动端振动 */
export function playRingtone() {
  startLoop((c) => {
    const now = c.currentTime;
    beep(c, [660], now + 0.02, 0.24, 0.16);
    beep(c, [880], now + 0.36, 0.24, 0.16);
    beep(c, [660], now + 0.70, 0.24, 0.16);
  }, 2000);

  // 移动端振动（循环）
  if (typeof navigator !== 'undefined' && navigator.vibrate) {
    const buzz = () => { try { navigator.vibrate?.([400, 200, 400]); } catch { /* ignore */ } };
    buzz();
    vibrateTimer = setInterval(buzz, 2000);
  }
}

export function stopRing() {
  if (timer) { clearInterval(timer); timer = null; }
  if (vibrateTimer) { clearInterval(vibrateTimer); vibrateTimer = null; }
  if (typeof navigator !== 'undefined' && navigator.vibrate) {
    try { navigator.vibrate?.(0); } catch { /* ignore */ }
  }
  // 不关闭 AudioContext（复用，避免重复解锁）；正在响的 osc 会自然结束
}
