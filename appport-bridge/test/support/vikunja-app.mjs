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

// Disposable Vikunja instance for the consumer test: the real Vikunja server, SQLite file in a temp dir, synthetic users,
// projects, tasks and API tokens created through Vikunja's own API. Nothing here is used by the bridge.
//
// VIKUNJA_BIN=/path/to/vikunja   use a prebuilt binary; otherwise `go build -tags osusergo .` runs at the repo root
//                                (as `mage build` does; needs frontend/dist/index.html, which mage/CI create).

import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { execFileSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import { randomBytes } from "node:crypto";
import { startProcess, freePort, tempDir, http, openOracle } from "../../lib/harness.mjs";

const REPO = join(dirname(fileURLToPath(import.meta.url)), "..", "..", "..");

function vikunjaBinary(work) {
  if (process.env.VIKUNJA_BIN) return process.env.VIKUNJA_BIN;
  const out = join(work, "vikunja");
  execFileSync("go", ["build", "-tags", "osusergo", "-o", out, "."], { cwd: REPO, stdio: "inherit" });
  return out;
}

const FULL = { projects: ["read_all", "read_one"], tasks: ["read_one", "update"] };
const READ_ONLY = { projects: ["read_all", "read_one"], tasks: ["read_one"] };

export async function startVikunja() {
  const work = tempDir("vikunja-appport-");
  const root = join(work, "root"); mkdirSync(root);
  const port = await freePort();
  const base = `http://127.0.0.1:${port}`;
  const env = {
    VIKUNJA_SERVICE_INTERFACE: `127.0.0.1:${port}`, VIKUNJA_SERVICE_ROOTPATH: root, VIKUNJA_SERVICE_PUBLICURL: `${base}/`,
    VIKUNJA_SERVICE_SECRET: randomBytes(24).toString("hex"), VIKUNJA_DATABASE_TYPE: "sqlite", VIKUNJA_DATABASE_PATH: join(root, "vikunja.db"),
    VIKUNJA_FILES_BASEPATH: join(root, "files"), VIKUNJA_SENTRY_ENABLED: "false", VIKUNJA_MAILER_ENABLED: "false", HOME: root
  };
  const proc = await startProcess(vikunjaBinary(work), ["web"], { env, readyUrl: `${base}/api/v1/info`, name: "vikunja", timeoutMs: 60_000 });

  const users = {};
  for (const [username, password] of [["bob", "synthetic-bob-pw-1"], ["carol", "synthetic-carol-pw-2"]]) {
    const reg = await http(base, "POST", "/api/v1/register", { body: { username, email: `${username}@example.invalid`, password } });
    if (reg.status !== 200) throw new Error(`register ${username}: ${reg.status}`);
    const jwt = (await http(base, "POST", "/api/v1/login", { body: { username, password } })).json.token;
    const token = async (permissions) => {
      const r = await http(base, "PUT", "/api/v1/tokens", { headers: { authorization: `Bearer ${jwt}` }, body: { title: "appport consumer test", expires_at: new Date(Date.now() + 86_400_000).toISOString(), permissions } });
      if (!r.json?.token) throw new Error(`token for ${username}: ${r.status}`);
      return `Bearer ${r.json.token}`;
    };
    users[username] = { username, jwt, credential: await token(FULL), readOnlyCredential: await token(READ_ONLY) };
  }

  const fixtures = {};
  const asUser = (who) => ({ authorization: `Bearer ${users[who].jwt}` });
  for (const who of ["bob", "carol"]) {
    const p = await http(base, "PUT", "/api/v1/projects", { headers: asUser(who), body: { title: `Synthetic ${who} garden` } });
    const t = await http(base, "PUT", `/api/v1/projects/${p.json.id}/tasks`, { headers: asUser(who), body: { title: `Water ${who}'s turnips`, description: "synthetic task", priority: 3 } });
    if (t.status !== 201 && t.status !== 200) throw new Error(`task for ${who}: ${t.status}`);
    fixtures[who] = { projectId: p.json.id, taskId: t.json.id };
  }

  const oracle = await openOracle(join(root, "vikunja.db"));
  return {
    base, users, fixtures, work,
    /** Independent observation: Vikunja's own `tasks` table, read directly (read-only). */
    observeTask: (id) => oracle.get("SELECT t.id, t.title, t.description, t.project_id, t.done, t.done_at, t.priority, u.username AS created_by FROM tasks t JOIN users u ON u.id = t.created_by_id WHERE t.id = ?", id),
    /** Stored projects `username` owns (the synthetic data shares nothing, so ownership is the whole visibility rule here). */
    observeOwnedProjects: (username) => oracle.all("SELECT p.id, p.title FROM projects p JOIN users u ON u.id = p.owner_id WHERE u.username = ? ORDER BY p.id", username),
    async stop() { oracle.close(); await proc.stop(); }
  };
}
