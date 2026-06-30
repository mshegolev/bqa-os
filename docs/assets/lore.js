// lore.js — in-browser, key-less LLM lore generation via Transformers.js.
//
// Runs a small instruction model fully client-side (WebGPU when available,
// WASM fallback). No API key, no backend — model weights are fetched once from
// the public Hugging Face CDN and cached in the browser. Used to write a short
// heroic backstory for a hero from its role / skills / MCP.

const MODEL = "onnx-community/gemma-3-270m-it-ONNX";
let _genPromise = null;

async function pickDevice() {
  try {
    if (navigator.gpu && (await navigator.gpu.requestAdapter())) return "webgpu";
  } catch (_) { /* no webgpu */ }
  return "wasm";
}

// Lazily build (and reuse) the text-generation pipeline.
async function getGenerator(onProgress) {
  if (_genPromise) return _genPromise;
  _genPromise = (async () => {
    const { pipeline } = await import("https://cdn.jsdelivr.net/npm/@huggingface/transformers");
    const device = await pickDevice();
    return pipeline("text-generation", MODEL, { dtype: "q4", device, progress_callback: onProgress });
  })().catch((e) => { _genPromise = null; throw e; });
  return _genPromise;
}

function extractReply(output) {
  const g = output && output[0] && output[0].generated_text;
  if (Array.isArray(g)) return (g[g.length - 1] && g[g.length - 1].content) || "";
  if (typeof g === "string") return g;
  return "";
}

// Generate ~2 sentences of fantasy backstory for a hero. onProgress receives
// Transformers.js progress events during the one-time model download.
export async function generateLore(hero, onProgress) {
  const gen = await getGenerator(onProgress);
  const skills = (hero.skills || []).join(", ") || "general QA";
  const mcp = (hero.mcp || []).join(", ") || "none";
  const messages = [
    { role: "system", content: "You are a fantasy bard. Reply with a vivid 2-sentence heroic backstory only. No preamble, no lists." },
    { role: "user", content: `Hero: ${hero.label}. Role: ${hero.sdlc || "citadel guardian"}. Skills: ${skills}. Allied MCP relays: ${mcp}. Write their legend defending the Citadel against orc bugs, CVEs and the Deadline Warlord.` },
  ];
  const out = await gen(messages, { max_new_tokens: 90, temperature: 0.9, top_p: 0.9, do_sample: true });
  return extractReply(out).trim();
}
