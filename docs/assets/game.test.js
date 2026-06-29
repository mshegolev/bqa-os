const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

function loadGame() {
  const context = {
    console,
    fetch: () => Promise.reject(new Error("network disabled in tests")),
    localStorage: {
      getItem: () => null,
      setItem: () => {},
    },
    sessionStorage: {
      getItem: () => null,
    },
    document: {
      addEventListener: () => {},
      getElementById: () => null,
      querySelector: () => null,
    },
    window: {},
  };
  vm.createContext(context);
  for (const file of ["game-logs.js", "game.js"]) {
    const source = fs.readFileSync(path.join(__dirname, file), "utf8");
    vm.runInContext(source, context, { filename: file });
  }
  return context;
}

function runJSON(context, source) {
  return JSON.parse(vm.runInContext(`JSON.stringify(${source})`, context));
}

function gameHTML() {
  return fs.readFileSync(path.join(__dirname, "..", "game.html"), "utf8");
}

function inlineGameCSS() {
  const match = gameHTML().match(/<style>\s*([\s\S]*?)\s*<\/style>/);
  assert.ok(match, "docs/game.html should keep game-specific CSS inline for the static page");
  return match[1];
}

function cssValue(css, selector, property) {
  const escaped = property.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const propertyPattern = new RegExp(`${escaped}\\s*:\\s*([^;]+)`);
  let value = "";
  for (const match of css.matchAll(/([^{}]+)\{([^}]*)\}/g)) {
    const selectors = match[1].split(",").map((item) => item.trim());
    if (!selectors.includes(selector)) continue;
    const propertyMatch = match[2].match(propertyPattern);
    if (propertyMatch) value = propertyMatch[1].trim();
  }
  return value;
}

function modeledGameStartLayout(css, viewportWidth) {
  const wrapInlinePadding = 36;
  const panelChromeWidth = 42;
  const stageChromeWidth = 22;
  // Representative title/buttons/armory/leaderboard/upload-hint stack height.
  const startContentHeight = 530;
  const wrapContentWidth = Math.min(viewportWidth, 1100) - wrapInlinePadding;
  const mainColumnWidth = viewportWidth >= 1000 ? wrapContentWidth - 300 - 16 : wrapContentWidth;
  const panelContentWidth = mainColumnWidth - panelChromeWidth;
  const stageInnerWidth = panelContentWidth - stageChromeWidth;
  const canvasHeight = Math.round(stageInnerWidth * 0.6);

  const stageDisplay = cssValue(css, ".stage", "display") || "block";
  const startPosition = cssValue(css, ".start", "position") || "static";
  const startPlacement = [
    cssValue(css, ".start", "place-content"),
    cssValue(css, ".start", "justify-content"),
  ].filter(Boolean).join(" ");
  const startOverflow = cssValue(css, ".start", "overflow") || "visible";
  const startSizesStage = stageDisplay.includes("grid") && startPosition !== "absolute";
  const stageInnerHeight = startSizesStage ? Math.max(canvasHeight, startContentHeight) : canvasHeight;
  const centeredOffset = startPosition === "absolute" && /center/.test(startPlacement)
    ? (stageInnerHeight - startContentHeight) / 2
    : 0;
  const startTop = 8 + centeredOffset;
  const startBottom = startTop + startContentHeight;
  const commandTop = 8 + stageInnerHeight + 8 + 12;

  return {
    startTop,
    startBottom,
    commandTop,
    startOverflow,
  };
}

function runBrowserLayoutCheck() {
  const script = String.raw`
import json
import os
import sys
from pathlib import Path

from playwright.sync_api import sync_playwright

groups = [
    {"selector": ".resbar .r", "label": "HUD chip"},
    {"selector": "#mute", "label": "mute button"},
    {"selector": "#fs", "label": "fullscreen button"},
    {"selector": "#start h2", "label": "start title"},
    {"selector": "#lb-start .lb-row", "label": "leaderboard row"},
    {"selector": "#start .hint", "label": "upload hint"},
    {"selector": "#commands .cmd", "label": "command button"},
]
viewports = [
    {"name": "desktop", "width": 1100, "height": 1000},
    {"name": "mobile", "width": 390, "height": 1000},
]

def overlaps(a, b):
    return (
        a["left"] < b["right"] - 0.5
        and a["right"] > b["left"] + 0.5
        and a["top"] < b["bottom"] - 0.5
        and a["bottom"] > b["top"] + 0.5
    )

def collect_rects(page):
    return page.evaluate(
        """(groups) => groups.flatMap((group) =>
            Array.from(document.querySelectorAll(group.selector)).map((node, index) => {
                const rect = node.getBoundingClientRect();
                const style = window.getComputedStyle(node);
                return {
                    label: group.label + " " + (index + 1),
                    selector: group.selector,
                    visible: style.display !== "none" && style.visibility !== "hidden" && rect.width > 0 && rect.height > 0,
                    left: rect.left,
                    right: rect.right,
                    top: rect.top,
                    bottom: rect.bottom,
                    width: rect.width,
                    height: rect.height,
                };
            })
        )""",
        groups,
    )

html_url = Path(os.environ["BQA_GAME_HTML"]).resolve().as_uri()
failures = []
with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page()
    for viewport in viewports:
        page.set_viewport_size({"width": viewport["width"], "height": viewport["height"]})
        page.goto(html_url, wait_until="load")
        page.wait_for_function(
            "document.querySelectorAll('#commands .cmd').length >= 9 && "
            "document.querySelectorAll('#lb-start .lb-row').length >= 3"
        )
        rects = [rect for rect in collect_rects(page) if rect["visible"]]
        for left_index, left in enumerate(rects):
            for right in rects[left_index + 1:]:
                if overlaps(left, right):
                    failures.append({
                        "viewport": viewport["name"],
                        "first": left,
                        "second": right,
                    })
    browser.close()

if failures:
    print(json.dumps(failures, indent=2))
    sys.exit(1)
`;
  return spawnSync("python3", ["-c", script], {
    cwd: path.join(__dirname, ".."),
    encoding: "utf8",
    env: {
      ...process.env,
      BQA_GAME_HTML: path.join(__dirname, "..", "game.html"),
    },
    timeout: 30000,
  });
}

test("Story Demo starts with only the feature worker recruit unlocked", () => {
  const context = loadGame();
  const states = runJSON(context, `(() => {
    resetDemoRecruitUnlocks();
    return demoRecruitSnapshot().map((state) => [state.id, state.locked]);
  })()`);

  assert.deepEqual(states, [
    ["feature_worker", false],
    ["prompt_smith", true],
    ["hardening_engineer", true],
    ["incident_defender", true],
    ["sentinel_archer", true],
  ]);
});

test("Story Demo unlocks recruitable units in teaching sequence", () => {
  const context = loadGame();
  const states = runJSON(context, `(() => {
    resetDemoRecruitUnlocks();
    unlockDemoRecruitsForEvent({ type: "feature_detected" });
    unlockDemoRecruitsForEvent({ type: "prompt_loaded" });
    unlockDemoRecruitsForEvent({ type: "hardening_rule_added" });
    unlockDemoRecruitsForEvent({ type: "incident_started" });
    return demoRecruitSnapshot().map((state) => [state.id, state.locked]);
  })()`);

  assert.deepEqual(states, [
    ["feature_worker", false],
    ["prompt_smith", false],
    ["hardening_engineer", false],
    ["incident_defender", false],
    ["sentinel_archer", false],
  ]);
});

test("Survival mode does not lock recruit commands", () => {
  const context = loadGame();
  const locked = runJSON(context, `(() => {
    resetDemoRecruitUnlocks();
    G.mode = "survival";
    G.state = "playing";
    return COMMANDS
      .filter((command) => command.kind === "recruit")
      .map((command) => [command.id, isCommandLocked(command)]);
  })()`);

  assert.deepEqual(locked, [
    ["feature_worker", false],
    ["prompt_smith", false],
    ["hardening_engineer", false],
    ["incident_defender", false],
    ["sentinel_archer", false],
  ]);
});

test("Story Demo renders locked recruit buttons with lock glyphs and hints", () => {
  const context = loadGame();
  const rendered = runJSON(context, `(() => {
    const buttons = {};
    function makeButton(command) {
      const children = {
        '.cmd-g': { textContent: command.glyph },
        '.cmd-cost': { textContent: '' },
      };
      return {
        dataset: { id: command.id },
        disabled: false,
        title: '',
        classList: {
          values: {},
          toggle(name, on) { this.values[name] = !!on; },
          contains(name) { return !!this.values[name]; },
        },
        querySelector(selector) { return children[selector] || null; },
      };
    }
    for (const command of COMMANDS) buttons[command.id] = makeButton(command);
    document.getElementById = (id) => id === 'commands' ? {
      querySelector(selector) {
        const match = selector.match(/data-id="([^"]+)"/);
        return match ? buttons[match[1]] : null;
      },
    } : null;

    resetDemoRecruitUnlocks();
    G.mode = 'demo';
    G.state = 'playing';
    G.res = { features: 99, prompts: 99, hardening: 99, logs: 99, trust: 60 };
    updateCommands();

    return DEMO_RECRUIT_UNLOCK_ORDER.map((id) => {
      const button = buttons[id];
      return {
        id,
        locked: button.classList.contains('locked'),
        disabled: button.disabled,
        glyph: button.querySelector('.cmd-g').textContent,
        hint: button.querySelector('.cmd-cost').textContent,
        title: button.title,
      };
    });
  })()`);

  assert.deepEqual(rendered, [
    { id: "feature_worker", locked: false, disabled: false, glyph: "★", hint: "8★", title: "" },
    { id: "prompt_smith", locked: true, disabled: true, glyph: "🔒", hint: "🔒 After first feature", title: "After first feature" },
    { id: "hardening_engineer", locked: true, disabled: true, glyph: "🔒", hint: "🔒 After prompt smithing", title: "After prompt smithing" },
    { id: "incident_defender", locked: true, disabled: true, glyph: "🔒", hint: "🔒 After hardening", title: "After hardening" },
    { id: "sentinel_archer", locked: true, disabled: true, glyph: "🔒", hint: "🔒 When incidents start", title: "When incidents start" },
  ]);
});

test("game start screen layout keeps title, leaderboard, upload hint, and command bar readable after resize", () => {
  const css = inlineGameCSS();
  const viewports = [
    { name: "desktop", width: 1100 },
    { name: "mobile", width: 390 },
  ];

  for (const viewport of viewports) {
    const layout = modeledGameStartLayout(css, viewport.width);

    assert.ok(
      layout.startTop >= 8,
      `${viewport.name} start title should not be vertically centered above the stage into the HUD`,
    );
    assert.ok(
      layout.startBottom <= layout.commandTop,
      `${viewport.name} start content should not overlap the command/recruit bar`,
    );
    assert.equal(
      layout.startOverflow,
      "auto",
      `${viewport.name} start content should scroll inside its own panel when viewport height is constrained`,
    );
  }
});

test(
  "browser layout has no visible HUD, start screen, leaderboard, upload hint, or command overlaps",
  { skip: process.env.BQA_BROWSER_QA !== "1" ? "set BQA_BROWSER_QA=1 to require local Playwright browser QA" : false },
  () => {
    const result = runBrowserLayoutCheck();

    assert.equal(
      result.status,
      0,
      [
        "browser overlap check failed",
        result.stdout.trim(),
        result.stderr.trim(),
      ].filter(Boolean).join("\n"),
    );
  },
);
