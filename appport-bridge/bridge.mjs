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

// Vikunja AppPort bridge: capabilities over Vikunja's own HTTP API, called with the caller's own Vikunja API token
// (`Authorization: Bearer tk_...`). Vikunja's routes, token scopes, permission checks (Project.CanRead, Task.CanRead,
// Task.CanUpdate) and storage decide everything; this process stores nothing.
//
//   APPPORT_BRIDGE_UPSTREAM=http://127.0.0.1:3456 [APPPORT_BRIDGE_PORT=4420] [APPPORT_BRIDGE_PUBLIC=1] node bridge.mjs

import { defineCapability, s } from "@appport/sdk";
import { startBridge, callApplication, runBridgeMain } from "./lib/bridge-kit.mjs";

const READ = "vikunja.read";
const WRITE = "vikunja.tasks.write";
const ZERO = "0001-01-01T00:00:00Z"; // Vikunja's "unset" time
const time = (t) => (!t || t.startsWith("0001-01-01") ? null : new Date(t).toISOString());

const Project = s.object({
  id: s.integer({ description: "Negative ids are Vikunja's virtual projects (saved filters), not stored projects" }),
  title: s.string(),
  description: s.string(),
  parentProjectId: s.integer(),
  isArchived: s.boolean(),
  owner: s.nullable(s.string()),
  virtual: s.boolean()
});
const Task = s.object({
  id: s.integer(), title: s.string(), description: s.string(), projectId: s.integer(),
  done: s.boolean(), doneAt: s.nullable(s.string()), dueDate: s.nullable(s.string()),
  priority: s.integer(), percentDone: s.number(), identifier: s.string(),
  createdBy: s.nullable(s.string()), created: s.string(), updated: s.string()
});
const project = (p) => ({ id: p.id, title: p.title ?? "", description: p.description ?? "", parentProjectId: p.parent_project_id ?? 0, isArchived: Boolean(p.is_archived), owner: p.owner?.username ?? null, virtual: p.id < 0 });
const task = (t) => ({
  id: t.id, title: t.title ?? "", description: t.description ?? "", projectId: t.project_id, done: Boolean(t.done), doneAt: time(t.done_at),
  dueDate: time(t.due_date), priority: t.priority ?? 0, percentDone: t.percent_done ?? 0, identifier: t.identifier ?? "",
  createdBy: t.created_by?.username ?? null, created: new Date(t.created).toISOString(), updated: new Date(t.updated).toISOString()
});
const TaskId = s.integer({ minimum: 1, maximum: Number.MAX_SAFE_INTEGER });

export function capabilities(upstream) {
  const authority = "urn:appport:vikunja";
  const resource = "urn:appport:vikunja:tasks";

  const listProjects = defineCapability({
    name: "vikunja.projects.list", version: 1,
    description: "List the projects Vikunja lets the caller see.",
    input: s.object({ includeArchived: s.optional(s.boolean()) }),
    output: s.object({ projects: s.array(Project) }),
    effect: "observation",
    authorization: [READ], authorizationContract: { required: true, scopes: [READ] },
    integration: { purpose: "List projects visible to the caller.", authority, approval: { required: false }, access: [{ resource: "urn:appport:vikunja:projects", mode: "reads" }] },
    handler: async ({ includeArchived }, ctx) => {
      // Vikunja's paginated v2 listing; every page is read so the result is the full list Vikunja shows the caller.
      const projects = [];
      for (let page = 1; page <= 50; page++) {
        const body = await callApplication(ctx, upstream, { path: "/api/v2/projects", query: { is_archived: includeArchived ? "true" : undefined, per_page: 200, page } });
        projects.push(...(body?.items ?? []).map(project));
        if (!body || page >= (body.total_pages ?? 1)) break;
      }
      return { projects };
    }
  });

  const getTask = defineCapability({
    name: "vikunja.tasks.get", version: 1,
    description: "Retrieve one task, including its done status, if Vikunja lets the caller read it.",
    input: s.object({ id: TaskId }),
    output: Task,
    effect: "observation",
    authorization: [READ], authorizationContract: { required: true, scopes: [READ] },
    integration: { purpose: "Read a task and its status.", authority, approval: { required: false }, access: [{ resource, mode: "reads" }] },
    handler: async ({ id }, ctx) => task(await callApplication(ctx, upstream, { path: `/api/v2/tasks/${id}` }))
  });

  const setDone = defineCapability({
    name: "vikunja.tasks.setdone", version: 1,
    description: "Mark a task done or not done. Vikunja checks that the caller may update the task.",
    input: s.object({ id: TaskId, done: s.boolean() }),
    output: Task,
    effect: "consequential",
    authorization: [WRITE], authorizationContract: { required: true, scopes: [WRITE] },
    idempotency: "unsupported",
    integration: {
      purpose: "Set a task's done flag.", authority, approval: { required: false },
      effects: [{ type: "urn:appport:vikunja#task-done-changed", description: "The task's done flag (and Vikunja's done_at) change." }],
      access: [{ resource, mode: "changes", operations: ["update"] }],
      verify: [{ id: "task-done-stored", kind: "state", via: { capability: "vikunja.tasks.get", version: 1, input: { id: { $input: "/id" } } }, expect: [{ at: "/done", equals: { $input: "/done" } }] }],
      failures: [
        { code: "FORBIDDEN", retry: "unsafe", effect: "none", description: "Vikunja refused: the caller may not update this task." },
        { code: "NOT_FOUND", retry: "unsafe", effect: "none" },
        { code: "INTERNAL_ERROR", retry: "after-verification", effect: "unknown", description: "Read the task before trying again." }
      ]
    },
    // PATCH on /api/v2 updates only the fields sent and runs Vikunja's normal update path (Task.CanUpdate, done_at handling).
    // (The frozen v1 POST /tasks/{id} is a full replace and would clear fields that are not sent.) All calls use /api/v2.
    handler: async ({ id, done }, ctx) => task(await callApplication(ctx, upstream, { method: "PATCH", path: `/api/v2/tasks/${id}`, body: { done } }))
  });

  return [listProjects, getTask, setDone];
}

export async function build(config) {
  const upstream = { baseUrl: config.upstream, headers: (credential) => ({ authorization: credential }) };
  return startBridge({
    application: { id: "io.vikunja", name: "Vikunja", version: "appport-bridge-1", description: "Vikunja projects and tasks, through Vikunja's own API with the caller's own API token." },
    capabilities: capabilities(upstream),
    resources: [
      { id: "urn:appport:vikunja:projects", kind: "collection", owner: "self", description: "Projects stored by this Vikunja instance." },
      { id: "urn:appport:vikunja:tasks", kind: "collection", owner: "self", description: "Tasks stored by this Vikunja instance." }
    ],
    scopes: [READ, WRITE],
    publish: config.publish,
    publishCapabilities: [
      { name: "vikunja.projects.list", version: 1, purpose: true },
      { name: "vikunja.tasks.get", version: 1, purpose: true },
      { name: "vikunja.tasks.setdone", version: 1, purpose: true }
    ],
    host: config.host, port: config.port
  });
}

if (import.meta.url === `file://${process.argv[1]}`) await runBridgeMain(build);
