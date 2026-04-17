/**
 * 消息提示音
 * 方案：Base64 内嵌短音频 + HTMLAudioElement 播放
 * 比 Web Audio API 更兼容 Chrome 的自动播放策略
 */

// 生成一个极短的 WAV 格式提示音（叮咚效果）
function generateNotificationWav(): string {
  const sampleRate = 22050;
  const duration = 0.25;
  const samples = Math.floor(sampleRate * duration);
  const buffer = new Float32Array(samples);

  for (let i = 0; i < samples; i++) {
    const t = i / sampleRate;
    // 双音叮咚：880Hz + 1100Hz，快速衰减
    const env1 = Math.max(0, 1 - t * 8);
    const env2 = Math.max(0, (t > 0.08 ? 1 - (t - 0.08) * 6 : 0));
    buffer[i] = Math.sin(2 * Math.PI * 880 * t) * 0.3 * env1
              + Math.sin(2 * Math.PI * 1100 * t) * 0.2 * env2;
  }

  // Encode as 16-bit PCM WAV
  const numChannels = 1;
  const bitsPerSample = 16;
  const byteRate = sampleRate * numChannels * (bitsPerSample / 8);
  const blockAlign = numChannels * (bitsPerSample / 8);
  const dataSize = samples * blockAlign;
  const headerSize = 44;
  const wav = new ArrayBuffer(headerSize + dataSize);
  const view = new DataView(wav);

  // WAV header
  const writeStr = (offset: number, s: string) => { for (let i = 0; i < s.length; i++) view.setUint8(offset + i, s.charCodeAt(i)); };
  writeStr(0, 'RIFF');
  view.setUint32(4, 36 + dataSize, true);
  writeStr(8, 'WAVE');
  writeStr(12, 'fmt ');
  view.setUint32(16, 16, true);
  view.setUint16(20, 1, true); // PCM
  view.setUint16(22, numChannels, true);
  view.setUint32(24, sampleRate, true);
  view.setUint32(28, byteRate, true);
  view.setUint16(32, blockAlign, true);
  view.setUint16(34, bitsPerSample, true);
  writeStr(36, 'data');
  view.setUint32(40, dataSize, true);

  // Write samples
  for (let i = 0; i < samples; i++) {
    const s = Math.max(-1, Math.min(1, buffer[i]));
    view.setInt16(headerSize + i * 2, s * 0x7FFF, true);
  }

  // Convert to base64 data URL
  const bytes = new Uint8Array(wav);
  let binary = '';
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]);
  return 'data:audio/wav;base64,' + btoa(binary);
}

let audioUrl: string | null = null;
let audioElement: HTMLAudioElement | null = null;

function getAudio(): HTMLAudioElement {
  if (!audioElement) {
    audioUrl = generateNotificationWav();
    audioElement = new Audio(audioUrl);
    audioElement.volume = 0.5;
  }
  return audioElement;
}

// 预加载：用户首次交互时初始化 audio 元素
if (typeof window !== 'undefined') {
  const init = () => {
    getAudio();
    ['click', 'keydown', 'touchstart', 'pointerdown'].forEach(e =>
      window.removeEventListener(e, init, true)
    );
  };
  ['click', 'keydown', 'touchstart', 'pointerdown'].forEach(e =>
    window.addEventListener(e, init, { capture: true, passive: true })
  );
}

/**
 * 播放消息提示音
 */
export function playMessageSound() {
  if (typeof window === 'undefined') return;
  try {
    const audio = getAudio();
    audio.currentTime = 0;
    audio.play().catch(() => { /* blocked by browser, silent */ });
  } catch { /* silent */ }
}

/**
 * 触发浏览器振动（移动端）
 */
export function vibrate() {
  try {
    if (typeof navigator !== 'undefined' && navigator.vibrate) {
      navigator.vibrate(200);
    }
  } catch { /* not supported */ }
}
