// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Independent AppPort consumer test for the Vikunja bridge.
// Real Vikunja server + real bridge process; the consumer uses only published @appport packages. Every success is
// checked against Vikunja's own database, read directly; nothing is accepted on the bridge's word.

import { test, before, after } from "node:test";
import assert from "node:assert/strict";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { existsSync } from "node:fs";
import { getIntegration } from "@appport/protocol";
import { planVerification, evaluateVerification } from "@appport/integration";
import { startProcess, freePort, removeDir } from "../lib/harness.mjs";
import { discover, connect, invoke, verdict, losingFetch, isCallTo } from "../lib/consumer.mjs";
import { startVikunja } from "./support/vikunja-app.mjs";

const BRIDGE = join(dirname(fileURLToPath(import.meta.url)), "..", "bridge.mjs");
let app, bridge, origin;
const startBridge = async (publish) => startProcess(process.execPath, [BRIDGE], { env: { APPPORT_BRIDGE_UPSTREAM: app.base, APPPORT_BRIDGE_PORT: String(await freePort()), APPPORT_BRIDGE_PUBLIC: publish ? "1" : "" }, readyLine: true, name: "bridge" });

before(async () => { app = await startVikunja(); bridge = await startBridge(true); origin = bridge.ready.url; });
after(async () => { await bridge?.stop(); await app?.stop(); if (app) removeDir(app.work); });

/** Runs the contract's declared read-back (published @appport/integration) through AppPort with a read credential. */
async function declaredVerification(name, input, output, credential) {
  const probe = await connect(origin, credential);
  try {
    const methods = getIntegration(await probe.load())?.capabilities?.[`${name}@1`]?.verify ?? [];
    const observations = {};
    for (const p of planVerification(methods, { input, output })) {
      try { observations[p.methodId] = { status: "ran", output: await probe.call(p.capability, p.input, { version: p.version }) }; }
      catch (e) { observations[p.methodId] = { status: "failed", errorCode: e?.code ?? "ERROR" }; }
    }
    return evaluateVerification(methods, { input, output }, observations, { application: "io.vikunja", capability: { name, version: 1 } }).outcome;
  } finally { await probe.close(); }
}

test("the bridge refuses to start unless explicitly configured", async () => {
  await assert.rejects(startProcess(process.execPath, [BRIDGE], { env: { APPPORT_BRIDGE_UPSTREAM: "" }, readyLine: true, timeoutMs: 5000, name: "unconfigured" }), /APPPORT_BRIDGE_UPSTREAM is not set/);
});

test("discovery is opt-in: without APPPORT_BRIDGE_PUBLIC nothing is published", async () => {
  const quiet = await startBridge(false);
  try { assert.equal((await discover(quiet.ready.url)).state, "not-published"); } finally { await quiet.stop(); }
});

test("discovery: reference-verified; two reads and one scoped mutation; nothing established", async () => {
  const found = await discover(origin);
  assert.equal(found.state, "reference-verified", found.reasons.join());
  const caps = Object.fromEntries(found.capabilities.map((c) => [`${c.name}@${c.version}`, c]));
  assert.deepEqual(Object.keys(caps).sort(), ["vikunja.projects.list@1", "vikunja.tasks.get@1", "vikunja.tasks.setdone@1"]);
  assert.equal(caps["vikunja.projects.list@1"].declared.effect, "observation");
  assert.equal(caps["vikunja.tasks.get@1"].declared.effect, "observation");
  assert.equal(caps["vikunja.tasks.setdone@1"].declared.effect, "consequential");
  assert.deepEqual(caps["vikunja.tasks.setdone@1"].declared.authorization, { scopes: ["vikunja.tasks.write"] });
  assert.deepEqual(found.established, { publisherTrust: false, authorization: false, approval: false, invoked: false });
});

test("projects.list: exactly the stored projects Vikunja shows bob (independent DB check)", async () => {
  const client = await connect(origin, app.users.bob.credential);
  try {
    const out = await invoke(client, "vikunja.projects.list", {});
    const stored = out.output.projects.filter((p) => !p.virtual).map((p) => `${p.id}:${p.title}`).sort();
    const expected = app.observeOwnedProjects("bob").map((p) => `${p.id}:${p.title}`).sort();
    assert.equal(verdict(out, JSON.stringify(stored) === JSON.stringify(expected)), "confirmed", `${stored} vs ${expected}`);
    assert.ok(!out.output.projects.some((p) => p.id === app.fixtures.carol.projectId), "carol's project must not appear");
    assert.ok(out.output.projects.filter((p) => p.virtual).every((p) => p.id < 0), "virtual projects are Vikunja's saved filters, flagged as such");
  } finally { await client.close(); }
});

test("tasks.get: task and status confirmed against Vikunja's database", async () => {
  const client = await connect(origin, app.users.bob.credential);
  try {
    const out = await invoke(client, "vikunja.tasks.get", { id: app.fixtures.bob.taskId });
    const row = app.observeTask(app.fixtures.bob.taskId);
    const agrees = out.output.title === row.title && out.output.description === row.description && out.output.projectId === row.project_id
      && out.output.done === Boolean(row.done) && out.output.priority === row.priority && out.output.createdBy === row.created_by;
    assert.equal(verdict(out, agrees), "confirmed");
  } finally { await client.close(); }
});

test("authorization is Vikunja's: carol cannot read bob's task (FORBIDDEN), although it exists", async () => {
  const client = await connect(origin, app.users.carol.credential);
  try {
    const out = await invoke(client, "vikunja.tasks.get", { id: app.fixtures.bob.taskId });
    assert.equal(out.kind, "denied"); assert.equal(out.code, "FORBIDDEN");
    assert.equal(app.observeTask(app.fixtures.bob.taskId).created_by, "bob");
  } finally { await client.close(); }
});

test("no credential: refused by AppPort (UNAUTHORIZED); an unknown token is rejected by Vikunja", async () => {
  for (const [credential, code] of [[undefined, "UNAUTHORIZED"], ["Bearer tk_not-a-real-token", "UNAUTHORIZED"]]) {
    const client = await connect(origin, credential);
    try { const out = await invoke(client, "vikunja.projects.list", {}); assert.equal(out.kind, "denied"); assert.equal(out.code, code); }
    finally { await client.close(); }
  }
});

test("application-level failure: a task that does not exist is NOT_FOUND; malformed input is INVALID_INPUT", async () => {
  const client = await connect(origin, app.users.bob.credential);
  try {
    const missing = await invoke(client, "vikunja.tasks.get", { id: 999999 });
    assert.equal(verdict(missing), "failed"); assert.equal(missing.code, "NOT_FOUND");
    assert.equal(app.observeTask(999999), undefined);
    assert.equal((await invoke(client, "vikunja.tasks.get", { id: -1 })).code, "INVALID_INPUT");
    assert.equal((await invoke(client, "vikunja.tasks.setdone", { id: app.fixtures.bob.taskId, done: "yes" })).code, "INVALID_INPUT");
  } finally { await client.close(); }
});

// ---------------------------------------------------------------- the one mutation, tested separately

test("setdone: confirmed by the declared read-back AND by Vikunja's database; other fields untouched", async () => {
  const id = app.fixtures.bob.taskId;
  const before = app.observeTask(id);
  assert.equal(before.done, 0);
  const client = await connect(origin, app.users.bob.credential);
  try {
    const out = await invoke(client, "vikunja.tasks.setdone", { id, done: true });
    assert.equal(out.kind, "response");
    const declared = await declaredVerification("vikunja.tasks.setdone", { id, done: true }, out.output, app.users.bob.readOnlyCredential);
    const after = app.observeTask(id);
    const dbAgrees = after.done === 1 && after.done_at !== null && after.title === before.title && after.description === before.description && after.priority === before.priority;
    assert.equal(declared, "verified");
    assert.equal(verdict(out, dbAgrees), "confirmed");
    // and back, so later tests start from a known state
    const undo = await invoke(client, "vikunja.tasks.setdone", { id, done: false });
    assert.equal(verdict(undo, app.observeTask(id).done === 0), "confirmed");
  } finally { await client.close(); }
});

test("setdone denied: carol cannot change bob's task; the database is unchanged", async () => {
  const id = app.fixtures.bob.taskId;
  const before = app.observeTask(id);
  const client = await connect(origin, app.users.carol.credential);
  try {
    const out = await invoke(client, "vikunja.tasks.setdone", { id, done: true });
    assert.equal(verdict(out), "denied"); assert.equal(out.code, "FORBIDDEN");
    assert.deepEqual(app.observeTask(id), before);
  } finally { await client.close(); }
});

test("setdone with a read-only Vikunja token: Vikunja's token scopes refuse it; the database is unchanged", async () => {
  const id = app.fixtures.bob.taskId;
  const before = app.observeTask(id);
  const client = await connect(origin, app.users.bob.readOnlyCredential);
  try {
    const out = await invoke(client, "vikunja.tasks.setdone", { id, done: true });
    assert.equal(out.kind, "denied"); assert.equal(out.code, "UNAUTHORIZED", "Vikunja answers 401 for a token without the scope");
    assert.deepEqual(app.observeTask(id), before);
  } finally { await client.close(); }
});

test("setdone with a lost response: indeterminate, not retried, and only the database shows what happened", async () => {
  const id = app.fixtures.carol.taskId;
  assert.equal(app.observeTask(id).done, 0);
  let sends = 0;
  const lose = losingFetch(isCallTo("vikunja.tasks.setdone"));
  const client = await connect(origin, app.users.carol.credential, { fetch: async (u, i) => { if (isCallTo("vikunja.tasks.setdone")(String(u), i)) sends++; return lose(u, i); } });
  try {
    const out = await invoke(client, "vikunja.tasks.setdone", { id, done: true });
    assert.equal(out.kind, "no-response");
    assert.equal(verdict(out, true), "indeterminate", "a lost answer is never reported as confirmed, even if the effect is later observed");
    assert.equal(sends, 1, "the mutation was sent exactly once and not retried");
    assert.equal(app.observeTask(id).done, 1, "the operation did run: assuming 'not performed' and retrying blindly would have been wrong");
  } finally { await client.close(); }
});

test("cleanup: stopping removes the disposable Vikunja data", async () => {
  const work = app.work;
  await bridge.stop(); await app.stop(); removeDir(work);
  assert.equal(existsSync(work), false);
  bridge = null; app = null;
});
