// マイク入力の波形からピッチ(基本周波数)を検出する共通ロジック。
// pitch-test.html(検証用プロトタイプ)とindex.html(本編、Linkの移動操作)
// の両方から利用する。

// computeRMS は波形データの実効値(音量の大きさ)を計算する。
function computeRMS(buf) {
  let sum = 0;
  for (let i = 0; i < buf.length; i++) {
    sum += buf[i] * buf[i];
  }
  return Math.sqrt(sum / buf.length);
}

// rmsToDb はRMSをデシベルに変換する(0dB = 振幅1相当、値が小さいほど負に大きくなる)。
function rmsToDb(rms) {
  return 20 * Math.log10(Math.max(rms, 1e-8));
}

// 自己相関法で波形の基本周波数(ピッチ)を推定する。
// 音量による足切り(音量が小さければ無視する)は呼び出し側で行う。
function detectPitch(buf, sampleRate) {
  const size = buf.length;

  // 振幅が閾値を下回る位置で波形の前後をトリムし、無音部分の影響を減らす。
  const threshold = 0.2;
  let start = 0;
  for (let i = 0; i < size / 2; i++) {
    if (Math.abs(buf[i]) < threshold) {
      start = i;
      break;
    }
  }
  let end = size - 1;
  for (let i = 1; i < size / 2; i++) {
    if (Math.abs(buf[size - i]) < threshold) {
      end = size - i;
      break;
    }
  }

  const trimmed = buf.slice(start, end);
  const n = trimmed.length;
  if (n < 2) {
    return -1;
  }

  const correlation = new Float32Array(n);
  for (let lag = 0; lag < n; lag++) {
    let sum = 0;
    for (let i = 0; i < n - lag; i++) {
      sum += trimmed[i] * trimmed[i + lag];
    }
    correlation[lag] = sum;
  }

  // 最初の下り坂を抜けた後の最大ピークを、基本周期の候補として探す。
  let d = 0;
  while (d < n - 1 && correlation[d] > correlation[d + 1]) {
    d++;
  }
  let maxValue = -1;
  let maxLag = -1;
  for (let i = d; i < n; i++) {
    if (correlation[i] > maxValue) {
      maxValue = correlation[i];
      maxLag = i;
    }
  }
  if (maxLag <= 0) {
    return -1;
  }

  // 放物線補間でピーク位置を細かく補正する。
  let period = maxLag;
  const x1 = correlation[maxLag - 1] || 0;
  const x2 = correlation[maxLag];
  const x3 = correlation[maxLag + 1] || 0;
  const a = (x1 + x3 - 2 * x2) / 2;
  const b = (x3 - x1) / 2;
  if (a !== 0) {
    period = maxLag - b / (2 * a);
  }

  return sampleRate / period;
}

function frequencyToNoteName(freq) {
  const NOTE_NAMES = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B'];
  const noteNumber = 12 * Math.log2(freq / 440) + 69;
  const rounded = Math.round(noteNumber);
  const name = NOTE_NAMES[((rounded % 12) + 12) % 12];
  const octave = Math.floor(rounded / 12) - 1;
  return name + octave;
}
