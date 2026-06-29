const assert = require("node:assert/strict");
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
