import http from 'k6/http';
import encoding from 'k6/encoding';
import exec from 'k6/execution';
import { SharedArray } from 'k6/data';
import { Counter, Rate, Trend } from 'k6/metrics';
import { WebSocket } from 'k6/websockets';

function required(name) {
  if (!__ENV[name]) throw new Error(`${name} is required; see deploy/loadtest/README.md`);
  return __ENV[name];
}

function number(name, fallback, min, max = Infinity) {
  const value = Number(__ENV[name] === undefined ? fallback : __ENV[name]);
  if (!Number.isFinite(value) || value < min || value > max) {
    throw new Error(`${name} must be between ${min} and ${max}`);
  }
  return value;
}

if (required('LOAD_CONFIRM') !== 'staging') throw new Error('LOAD_CONFIRM must be staging');
const baseURL = required('BASE_URL').replace(/\/+$/, '');
if (!/^https?:\/\/[^/?#@]+(?:\/[^?#]*)?$/.test(baseURL)) throw new Error('BASE_URL must be an HTTP(S) application URL');
const runID = required('RUN_ID');
if (!/^[A-Za-z0-9_-]{1,64}$/.test(runID)) throw new Error('RUN_ID must contain 1-64 letters, digits, underscores or hyphens');
const users = number('USERS', 20000, 2);
if (!Number.isInteger(users)) throw new Error('USERS must be an integer');
const activeRatio = number('ACTIVE_RATIO', 0.2, 0, 1);
const actionRate = number('ACTIONS_PER_SECOND', 0.2, 0.001, 100);
const editRatio = number('EDIT_RATIO', 0.2, 0, 1);
const probeRatio = number('PROBE_RATIO', 0.1, 0, 1);
if (editRatio + probeRatio > 1) throw new Error('EDIT_RATIO + PROBE_RATIO must not exceed 1');
const rampSeconds = number('RAMP_SECONDS', 300, 1);
const holdSeconds = number('HOLD_SECONDS', 600, 1);
const downSeconds = number('DOWN_SECONDS', 60, 1);
const probeTimeout = number('PROBE_TIMEOUT_SECONDS', 10, 0.1) * 1000;
const httpTimeout = number('HTTP_TIMEOUT_SECONDS', 10, 0.1) * 1000;
const authTimeout = number('AUTH_TIMEOUT_SECONDS', 5, 0.1) * 1000;
const drainTime = Math.max(probeTimeout, httpTimeout) + 1000;
const sessionSeconds = number('SESSION_SECONDS', 300, (drainTime + 1000) / 800);
const graceSeconds = Math.ceil(sessionSeconds * 1.2 + httpTimeout / 1000 + 1);
const fixturePath = required('FIXTURES');

const fixtures = new SharedArray('vikunja-load-users', () => {
  const entries = JSON.parse(open(fixturePath));
  if (!Array.isArray(entries) || entries.length < users) throw new Error(`FIXTURES needs at least ${users} entries`);
  const principals = new Set();
  const tasks = new Set();
  const expiresAfter = Date.now() / 1000 + rampSeconds + holdSeconds + downSeconds + graceSeconds;
  for (let index = 0; index < entries.length; index++) {
    const entry = entries[index];
    let claims;
    try {
      claims = JSON.parse(encoding.b64decode(entry.token.split('.')[1], 'rawurl', 's'));
    } catch (_) {
      throw new Error(`Fixture ${index} needs a user JWT`);
    }
    if (claims.type !== 1 || !Number.isSafeInteger(claims.id) || claims.id <= 0 || principals.has(claims.id)) {
      throw new Error(`Fixture ${index} must have a distinct user principal`);
    }
    if (!Number.isFinite(claims.exp) || claims.exp <= expiresAfter) throw new Error(`Fixture ${index} JWT expires before test completion`);
    if (!Number.isSafeInteger(entry.taskId) || entry.taskId <= 0 || tasks.has(entry.taskId) ||
        !Number.isSafeInteger(entry.projectId) || entry.projectId <= 0) {
      throw new Error(`Fixture ${index} needs a distinct taskId and a valid projectId`);
    }
    principals.add(claims.id);
    tasks.add(entry.taskId);
  }
  return entries;
});

const authSuccess = new Rate('ws_auth_success');
const authDuration = new Trend('ws_auth_duration', true);
const authenticatedSessions = new Counter('ws_authenticated_sessions');
const unexpectedClose = new Rate('ws_unexpected_close');
const apiSuccess = new Rate('api_success');
const actionSkipped = new Counter('actions_skipped');
const deliverySuccess = new Rate('delivery_success');
const deliveryLatency = new Trend('delivery_latency', true);
const deliveryProbes = new Counter('delivery_probes');
const wrongRecipient = new Counter('wrong_recipient_probes');

const thresholds = {
  ws_auth_success: ['rate>0.999'],
  ws_auth_duration: ['p(95)<1000', 'p(99)<3000'],
  ws_authenticated_sessions: ['count>0'],
  ws_unexpected_close: [`rate<=${number('MAX_UNEXPECTED_CLOSE_RATE', 0, 0, 1)}`],
  wrong_recipient_probes: ['count==0'],
};
if (activeRatio > 0) {
  thresholds.api_success = ['rate>0.999'];
  thresholds['http_req_duration{workload:api}'] = ['p(95)<500', 'p(99)<1500'];
  thresholds.actions_skipped = ['count==0'];
  if (probeRatio > 0) {
    thresholds.delivery_probes = ['count>0'];
    thresholds.delivery_success = ['rate>0.999'];
    thresholds.delivery_latency = ['p(95)<1000', 'p(99)<3000'];
  }
}

export const options = {
  discardResponseBodies: true,
  systemTags: ['status', 'method', 'name', 'scenario', 'expected_response', 'error_code'],
  tags: { run_id: runID },
  thresholds,
  scenarios: {
    connected_users: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: `${rampSeconds}s`, target: users },
        { duration: `${holdSeconds}s`, target: users },
        { duration: `${downSeconds}s`, target: 0 },
      ],
      gracefulRampDown: `${graceSeconds}s`,
      gracefulStop: `${graceSeconds}s`,
    },
  },
};

function request(method, path, principal, name, body, contentType = 'application/json') {
  return http.asyncRequest(method, `${baseURL}/api/v2${path}`, body === undefined ? null : JSON.stringify(body), {
    headers: { Authorization: `Bearer ${principal.token}`, 'Content-Type': contentType },
    tags: { workload: 'api', name },
    timeout: `${httpTimeout}ms`,
    redirects: 0,
  }).then(response => {
    const ok = response.status >= 200 && response.status < 300;
    apiSuccess.add(ok);
    return ok;
  });
}

export default function () {
  const index = exec.vu.idInTest - 1;
  if (index >= fixtures.length) exec.test.abort('No fixture for global VU ID; use coordinated execution segments and enough fixtures');
  const principal = fixtures[index];
  const actor = fixtures[(index + 1) % fixtures.length];
  const active = ((exec.vu.idInTest * 2654435761) % 1000000) / 1000000 < activeRatio;
  const prefix = `vkload:${runID}:${exec.vu.idInTest}:${exec.scenario.iterationInInstance}:`;
  const started = Date.now();
  const endsAt = started + sessionSeconds * 1000 * (0.8 + Math.random() * 0.4);
  let sequence = 0;

  return new Promise(resolve => {
    const pending = new Map();
    let authenticated = false;
    let plannedClose = false;
    let finished = false;
    let busy = false;
    let workTimer;
    let firstActionTimer;
    let authTimer;
    let endTimer;
    const socket = new WebSocket(`${baseURL.replace(/^http/, 'ws')}/api/v1/ws`, [], { tags: { name: 'websocket' } });
    wrongRecipient.add(0);
    actionSkipped.add(0);

    function settle(marker, success) {
      const probe = pending.get(marker);
      if (!probe) return;
      clearTimeout(probe.timer);
      pending.delete(marker);
      deliverySuccess.add(success);
      if (success) deliveryLatency.add(probe.receivedAt - probe.started);
    }

    function finish() {
      if (finished) return;
      finished = true;
      clearTimeout(authTimer);
      clearTimeout(endTimer);
      clearTimeout(firstActionTimer);
      clearInterval(workTimer);
      if (!authenticated) authSuccess.add(false);
      unexpectedClose.add(!plannedClose);
      for (const marker of pending.keys()) settle(marker, false);
      // Spread reconnects after a pod drain or a broker failure.
      setTimeout(resolve, plannedClose ? 0 : 500 + Math.random() * 1000);
    }

    function close(planned) {
      plannedClose = planned;
      socket.close();
    }

    async function action() {
      if (finished || !authenticated || Date.now() + drainTime >= endsAt) return;
      if (busy) {
        actionSkipped.add(1);
        return;
      }
      busy = true;
      try {
        const choice = Math.random();
        if (choice < probeRatio) {
          const marker = `${prefix}${sequence++}`;
          const probe = { started: Date.now(), accepted: false, receivedAt: 0 };
          probe.timer = setTimeout(() => settle(marker, false), probeTimeout);
          pending.set(marker, probe);
          deliveryProbes.add(1);
          const accepted = await request('POST', `/tasks/${principal.taskId}/comments`, actor, 'delivery_probe', { comment: `<p>${marker}</p>` });
          if (!accepted) settle(marker, false);
          else {
            probe.accepted = true;
            if (probe.receivedAt) settle(marker, true);
          }
        } else if (choice < probeRatio + editRatio) {
          await request('PATCH', `/tasks/${principal.taskId}`, principal, 'task_edit',
            { title: `Load ${runID} user ${exec.vu.idInTest} session ${exec.scenario.iterationInInstance} edit ${sequence++}` }, 'application/merge-patch+json');
        } else if (Math.random() < 0.5) {
          await request('GET', `/tasks/${principal.taskId}`, principal, 'task_read');
        } else {
          await request('GET', `/projects/${principal.projectId}`, principal, 'project_read');
        }
      } finally {
        busy = false;
      }
    }

    socket.onopen = () => socket.send(JSON.stringify({ action: 'auth', token: principal.token }));
    socket.onmessage = event => {
      const message = JSON.parse(event.data);
      if (message.action === 'auth.success' && message.success && !authenticated) {
        authenticated = true;
        clearTimeout(authTimer);
        authSuccess.add(true);
        authDuration.add(Date.now() - started);
        authenticatedSessions.add(1);
        socket.send(JSON.stringify({ action: 'subscribe', event: 'notification.created' }));
        if (active) {
          firstActionTimer = setTimeout(() => {
            action();
            workTimer = setInterval(action, 1000 / actionRate);
          }, 250 + Math.random() * 1000 / actionRate);
        }
        return;
      }
      if (message.error) {
        close(false);
        return;
      }
      if (message.event !== 'notification.created') return;
      const notification = message.data && message.data.notification;
      const text = notification && notification.comment && notification.comment.comment;
      if (typeof text !== 'string') return;
      const markers = text.match(/vkload:[A-Za-z0-9_-]+:\d+:\d+:\d+/g) || [];
      for (const marker of markers) {
        if (!marker.startsWith(`vkload:${runID}:`)) continue;
        const receiver = Number(marker.split(':')[2]);
        if (receiver !== exec.vu.idInTest) {
          wrongRecipient.add(1);
          continue;
        }
        const probe = pending.get(marker);
        if (!probe) continue;
        probe.receivedAt = Date.now();
        if (probe.accepted) settle(marker, true);
      }
    };
    socket.onclose = finish;
    socket.onerror = () => close(false);
    authTimer = setTimeout(() => close(false), authTimeout);
    endTimer = setTimeout(() => close(true), endsAt - Date.now());
  });
}
