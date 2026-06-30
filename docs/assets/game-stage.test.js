"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

const ROOT = __dirname;

function fakeElement(tag) {
  return {
    tagName: tag,
    children: [],
    dataset: {},
    style: {},
    textContent: "",
    className: "",
    disabled: false,
    appendChild(child) {
      this.children.push(child);
      return child;
    },
    replaceChildren(...children) {
      this.children = children;
    },
    addEventListener() {},
    querySelector() {
      return null;
    },
  };
}

function loadGame(upgrades = {}) {
  const elements = new Map();
  const storage = new Map();
  storage.set("bqa-citadel-upgrades", JSON.stringify(upgrades));

  const document = {
    createElement: fakeElement,
    createTextNode(text) {
      return { textContent: String(text) };
    },
    getElementById(id) {
      if (!elements.has(id)) elements.set(id, fakeElement("div"));
      return elements.get(id);
    },
    querySelector(selector) {
      if (selector.startsWith("#")) return this.getElementById(selector.slice(1));
      return fakeElement("div");
    },
    addEventListener() {},
  };

  const context = {
    console,
    document,
    localStorage: {
      getItem(key) {
        return storage.has(key) ? storage.get(key) : null;
      },
      setItem(key, value) {
        storage.set(key, String(value));
      },
    },
    location: { reload() {} },
    requestAnimationFrame() {},
    sessionStorage: { getItem() { return null; } },
    window: {},
  };
  context.window = context;
  vm.createContext(context);

  for (const file of ["game-logs.js", "game.js"]) {
    vm.runInContext(fs.readFileSync(path.join(ROOT, file), "utf8"), context, { filename: file });
  }
  vm.runInContext(`
    globalThis.__game = {
      G,
      ENEMY_TYPES,
      spawnSurvivalWave,
      startSurvival
    };
  `, context);

  return context.__game;
}

test("survival starts at the stage unlocked by persistent player level", () => {
  const firstLevel = loadGame();
  firstLevel.startSurvival(null);
  firstLevel.spawnSurvivalWave();

  assert.equal(firstLevel.G.playerLevel, 1);
  assert.equal(firstLevel.G.stage, 1);
  assert.equal(firstLevel.G.wave, 1);
  assert.equal(firstLevel.G.wavesSurvived, 0);

  const thirdLevel = loadGame({ level: 2 });
  thirdLevel.startSurvival(null);
  thirdLevel.spawnSurvivalWave();

  assert.equal(thirdLevel.G.playerLevel, 3);
  assert.equal(thirdLevel.G.stage, 3);
  assert.equal(thirdLevel.G.wave, 7);
  assert.equal(thirdLevel.G.wavesSurvived, 0);

  thirdLevel.spawnSurvivalWave();

  assert.equal(thirdLevel.G.wave, 8);
  assert.equal(thirdLevel.G.wavesSurvived, 1);
});

test("stage and level increase survival spawn density and enemy power", () => {
  const firstLevel = loadGame();
  firstLevel.startSurvival(null);
  firstLevel.spawnSurvivalWave();
  const firstEnemy = firstLevel.G.enemies[0];

  const thirdLevel = loadGame({ level: 2 });
  thirdLevel.startSurvival(null);
  thirdLevel.spawnSurvivalWave();
  const advancedEnemy = thirdLevel.G.enemies.find((enemy) => enemy.type === "bug_grunt");

  assert.ok(thirdLevel.G.enemies.length > firstLevel.G.enemies.length);
  assert.ok(advancedEnemy.maxhp > firstLevel.ENEMY_TYPES.bug_grunt.hp);
  assert.ok(advancedEnemy.dmg > firstEnemy.dmg);
});
