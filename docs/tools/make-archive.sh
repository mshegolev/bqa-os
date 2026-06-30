#!/usr/bin/env bash
# make-archive.sh — build an uploadable archive.zip for Citadel: The Release War.
#
# ONE command, no arguments. Auto-discovers your local AI-coding sessions
# (Claude Code / Codex / OpenCode / droid), turns them into archive.json, and
# packs a STORED archive.zip that the decoder page can open. Everything stays
# local — nothing is uploaded by this script.
#
#   curl -fsSL <pages-url>/tools/make-archive.sh | bash
#   # or:  bash make-archive.sh [EXTRA_DIR]
#
# Needs: python3 + zip (both standard on most dev machines).
set -euo pipefail

OUT_JSON="archive.json"
OUT_ZIP="archive.zip"

python3 - "${1:-}" <<'PY'
import json, os, sys, glob

home = os.path.expanduser("~")
# Where the AI coding tools keep their session artifacts.
roots = [
    (os.path.join(home, ".claude"), "claude"),
    (os.path.join(home, ".config", "claude"), "claude"),
    (os.path.join(home, ".codex"), "codex"),
    (os.path.join(home, ".config", "codex"), "codex"),
    (os.path.join(home, ".opencode"), "opencode"),
    (os.path.join(home, ".config", "opencode"), "opencode"),
    (os.path.join(home, ".factory"), "droid"),
    (".claude", "claude"), (".codex", "codex"), (".opencode", "opencode"),
]
extra = sys.argv[1] if len(sys.argv) > 1 and sys.argv[1] else None
if extra:
    roots.insert(0, (extra, "etl"))

DOMAINS = ("etl", "graphql", "api", "data_quality", "bugs", "prompts")
def guess_domain(text, path):
    t = (path + " " + text[:400]).lower()
    for d, kw in (("etl", "etl airflow spark hive dag parquet reconcil"),
                  ("graphql", "graphql resolver schema mutation"),
                  ("api", "rest api endpoint http openapi"),
                  ("data_quality", "schema drift null check duplicate checksum data quality"),
                  ("prompts", "prompt system message instruction"),
                  ("bugs", "bug traceback exception error fail regression")):
        if any(w in t for w in kw.split()):
            return d
    return "bugs"

def clean(s):
    return " ".join(s.split())[:2000]

# Skip obvious non-session noise (caches, configs, locks, binaries).
SKIP = ("cache", "node_modules", ".lock", "lockfile", "config-bundle",
        "settings", "package.json", "package-lock", "tsconfig", "/bin/", "/tmp/")
files = []
for root, tool in roots:
    if not os.path.isdir(root):
        continue
    for ext in ("jsonl", "md", "txt", "log", "json"):  # transcripts first
        for p in glob.glob(os.path.join(root, "**", "*." + ext), recursive=True):
            low = p.lower()
            if any(s in low for s in SKIP):
                continue
            try:
                if 200 <= os.path.getsize(p) <= 5_000_000:  # skip empty + huge
                    files.append((p, tool, os.path.getmtime(p)))
            except OSError:
                pass

files.sort(key=lambda x: x[2], reverse=True)   # newest first
files = files[:40]                              # cap the warband source

sessions = []
for i, (p, tool, _) in enumerate(files, 1):
    try:
        with open(p, encoding="utf-8", errors="replace") as f:
            raw = f.read(8000)
    except OSError:
        continue
    text = clean(raw)
    if not text:
        continue
    title = os.path.splitext(os.path.basename(p))[0][:60]
    sessions.append({"id": f"s-{i:04d}", "tool": tool,
                     "domain": guess_domain(text, p), "title": title, "text": text})

doc = {"archive": "my-qa-archive", "version": 1,
       "note": "Local AI-coding sessions. Sanitize before sharing publicly.",
       "sessions": sessions}
with open("archive.json", "w", encoding="utf-8") as f:
    json.dump(doc, f, indent=2, ensure_ascii=False)

if not sessions:
    print("No local AI sessions found. Pass a folder of .txt/.md logs: bash make-archive.sh ./sessions", file=sys.stderr)
print(f"sessions: {len(sessions)}")
PY

zip -0 -q "$OUT_ZIP" "$OUT_JSON"
echo "✓ built $OUT_ZIP — drop it on the decoder page."
