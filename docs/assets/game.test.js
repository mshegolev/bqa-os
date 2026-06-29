"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const test = require("node:test");
const vm = require("node:vm");

function loadGame() {
  const context = {
    console,
    document: {
      addEventListener() {},
      createElement() { return {}; },
      querySelector() { return null; },
    },
    window: {},
  };
  context.window = context;
  vm.createContext(context);
  vm.runInContext(fs.readFileSync("docs/assets/game.js", "utf8"), context);
  return context;
}

test("rankAgainstOfficial marks only a beaten leader as rank 1", () => {
  const game = loadGame();
  const official = [
    { name: "Leader", waves: 12 },
    { name: "Second", waves: 9 },
    { name: "Third", waves: 7 },
  ];

  assert.equal(game.rankAgainstOfficial(13, official), 1);
  assert.equal(game.rankAgainstOfficial(10, official), 2);
  assert.equal(game.rankAgainstOfficial(8, official), 3);
  assert.equal(game.rankAgainstOfficial(12, official), 2);
  assert.equal(game.rankAgainstOfficial(7, official), 0);
  assert.equal(game.rankAgainstOfficial(0, official), 0);
});

test("buildLeaderboardRows adds the animated challenger row only for rank 1", () => {
  const game = loadGame();
  const official = [
    { name: "Leader", waves: 12 },
    { name: "Second", waves: 9 },
    { name: "Third", waves: 7 },
  ];

  const leadRows = game.buildLeaderboardRows(official, {
    name: "You",
    waves: 13,
    officialRank: 1,
    takesLead: true,
  });
  assert.deepEqual(leadRows.map((r) => r.name), ["You", "Leader", "Second"]);
  assert.equal(leadRows[0].isFanfareLeader, true);

  const rankTwoRows = game.buildLeaderboardRows(official, {
    name: "You",
    waves: 10,
    officialRank: 2,
    takesLead: false,
  });
  assert.deepEqual(rankTwoRows.map((r) => r.name), ["Leader", "Second", "Third"]);
  assert.equal(rankTwoRows.some((r) => r.isFanfareLeader), false);
});
