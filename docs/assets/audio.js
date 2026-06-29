/* ============================================================
   CITADEL — procedural audio (WebAudio). No asset files: a tiny
   chiptune loop plus synthesized SFX and per-unit "voice" blips.
   Exposes window.AUDIO = { play, music, toggle, enabled }.
   Created lazily on first user gesture (autoplay policy).
   ============================================================ */

"use strict";

(function () {
  const AC = window.AudioContext || window.webkitAudioContext;
  let ctx = null, master = null, enabled = true, timer = null, step = 0;

  function ac() {
    if (!ctx && AC) { ctx = new AC(); master = ctx.createGain(); master.gain.value = 0.16; master.connect(ctx.destination); }
    if (ctx && ctx.state === "suspended") ctx.resume();
    return ctx;
  }
  function tone(freq, dur, type, when, vol) {
    const c = ac(); if (!c || !enabled || !freq) return;
    const o = c.createOscillator(), g = c.createGain();
    o.type = type || "square"; o.frequency.value = freq;
    const t = when || c.currentTime;
    g.gain.setValueAtTime(0.0001, t);
    g.gain.linearRampToValueAtTime(vol || 0.2, t + 0.01);
    g.gain.exponentialRampToValueAtTime(0.0008, t + dur);
    o.connect(g); g.connect(master); o.start(t); o.stop(t + dur + 0.03);
  }

  // light pentatonic loop: melody on 16ths, bass on quarters
  const MELODY = [440, 0, 523, 0, 587, 0, 523, 0, 440, 0, 392, 0, 330, 0, 392, 0];
  const BASS = [110, 110, 147, 147, 165, 147, 131, 110];
  function music(on) {
    const c = ac(); if (!c) return;
    if (on) {
      if (timer) return; step = 0;
      timer = setInterval(() => {
        if (!enabled) return;
        const m = MELODY[step % MELODY.length]; if (m) tone(m, 0.16, "square", 0, 0.07);
        if (step % 2 === 0) tone(BASS[(step / 2) % BASS.length], 0.22, "triangle", 0, 0.11);
        step++;
      }, 168);
    } else if (timer) { clearInterval(timer); timer = null; }
  }

  // per-unit-type "voice" pitch
  const VOICE = {
    builder: 262, feature_worker: 392, prompt_smith: 523, hardening_engineer: 330,
    context_logger: 440, incident_defender: 294, release_captain: 659,
  };

  function play(name) {
    const c = ac(); if (!c || !enabled) return;
    const now = c.currentTime;
    if (name.indexOf("select:") === 0) {
      const base = VOICE[name.slice(7)] || 349;
      tone(base, 0.07, "square", 0, 0.18); tone(base * 1.5, 0.07, "square", now + 0.08, 0.13);
    } else if (name === "command") { tone(660, 0.06, "square", 0, 0.15); }
    else if (name === "kill") { tone(220, 0.12, "sawtooth", 0, 0.18); tone(120, 0.14, "sawtooth", now + 0.05, 0.15); }
    else if (name === "wave") { tone(98, 0.32, "sawtooth", 0, 0.2); tone(146, 0.32, "sawtooth", now + 0.16, 0.17); }
    else if (name === "win") { [523, 659, 784, 1047].forEach((f, i) => tone(f, 0.18, "square", now + i * 0.12, 0.16)); music(false); }
    else if (name === "lose") { [392, 330, 262, 196].forEach((f, i) => tone(f, 0.26, "triangle", now + i * 0.14, 0.16)); music(false); }
  }

  function toggle() {
    enabled = !enabled;
    if (master) master.gain.value = enabled ? 0.16 : 0;
    return enabled;
  }

  window.AUDIO = { play, music, toggle, get enabled() { return enabled; } };
})();
