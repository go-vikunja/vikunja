import { pathToFileURL } from 'node:url';

const API = 'https://api.github.com';
const LABEL_REVIEW = 'waiting for review';
const LABEL_REPLY = 'waiting for reply';
const REVIEW_LABEL_SPEC = {
	name: LABEL_REVIEW,
	color: '0075ca',
	description: "Maintainers' turn: the author replied or pushed since the last request for changes",
};
const MAINTAINER_ASSOCIATIONS = new Set(['OWNER', 'MEMBER', 'COLLABORATOR']);

const PR_FIELDS = `
  number
  isDraft
  authorAssociation
  author { login __typename }
  labels(first: 50) { nodes { name } }
  timelineItems(last: 100, itemTypes: [PULL_REQUEST_COMMIT, HEAD_REF_FORCE_PUSHED_EVENT, ISSUE_COMMENT, PULL_REQUEST_REVIEW, LABELED_EVENT, UNLABELED_EVENT, READY_FOR_REVIEW_EVENT]) {
    nodes {
      __typename
      ... on PullRequestCommit { commit { committedDate } }
      ... on HeadRefForcePushedEvent { createdAt }
      ... on IssueComment { createdAt author { login } }
      ... on PullRequestReview { submittedAt state authorAssociation author { login } }
      ... on LabeledEvent { createdAt label { name } actor { __typename } }
      ... on UnlabeledEvent { createdAt label { name } }
      ... on ReadyForReviewEvent { createdAt }
    }
  }
`;

export function skipReason(pr) {
	if (pr.isDraft) return 'draft';
	if (MAINTAINER_ASSOCIATIONS.has(pr.authorAssociation)) return 'maintainer';
	if (pr.author?.__typename === 'Bot') return 'bot';
	return null;
}

// Timestamps are epoch ms; -Infinity means "no such event".
export function turnTimes(pr, { answeredNow = false } = {}) {
	const author = pr.author?.login;
	let ask = -Infinity;
	let answer = -Infinity;
	const bump = (current, time) => Math.max(current, Date.parse(time));

	for (const n of pr.timelineItems?.nodes ?? []) {
		switch (n.__typename) {
			case 'PullRequestReview': {
				const isMaintainer = MAINTAINER_ASSOCIATIONS.has(n.authorAssociation) && n.author?.login !== author;
				if (n.state === 'CHANGES_REQUESTED' && isMaintainer) ask = bump(ask, n.submittedAt);
				if (n.author?.login === author) answer = bump(answer, n.submittedAt);
				break;
			}
			case 'IssueComment':
				if (n.author?.login === author) answer = bump(answer, n.createdAt);
				break;
			case 'PullRequestCommit':
				answer = bump(answer, n.commit.committedDate);
				break;
			case 'HeadRefForcePushedEvent':
			case 'ReadyForReviewEvent':
				answer = bump(answer, n.createdAt);
				break;
			case 'LabeledEvent':
				if (n.label.name === LABEL_REPLY && n.actor?.__typename !== 'Bot') ask = bump(ask, n.createdAt);
				break;
			case 'UnlabeledEvent':
				if (n.label.name === LABEL_REPLY) answer = bump(answer, n.createdAt);
				break;
		}
	}

	if (answeredNow) answer = Math.max(answer, Date.now());
	return { ask, answer };
}

// Returns the turn label, or null when the PR is out of scope.
export function decideTurnLabel(pr, opts) {
	if (skipReason(pr)) return null;
	const { ask, answer } = turnTimes(pr, opts);
	return ask > answer ? LABEL_REPLY : LABEL_REVIEW;
}

async function api(path, { method = 'GET', body, allow404 = false } = {}) {
	const res = await fetch(`${API}${path}`, {
		method,
		headers: {
			Authorization: `Bearer ${process.env.GITHUB_TOKEN}`,
			Accept: 'application/vnd.github+json',
			'X-GitHub-Api-Version': '2022-11-28',
			'Content-Type': 'application/json',
		},
		body: body === undefined ? undefined : JSON.stringify(body),
	});
	if (allow404 && res.status === 404) return null;
	const text = await res.text();
	if (!res.ok) throw new Error(`${method} ${path}: HTTP ${res.status} ${text}`);
	return text ? JSON.parse(text) : null;
}

async function graphql(query, variables) {
	const json = await api('/graphql', { method: 'POST', body: { query, variables } });
	if (json.errors?.length) throw new Error(`GraphQL: ${JSON.stringify(json.errors)}`);
	return json.data;
}

async function fetchPr(owner, name, number) {
	const data = await graphql(
		`query($owner: String!, $name: String!, $number: Int!) {
			repository(owner: $owner, name: $name) { pullRequest(number: $number) { ${PR_FIELDS} } }
		}`,
		{ owner, name, number },
	);
	return [data.repository.pullRequest];
}

async function fetchOpenPrs(owner, name) {
	const prs = [];
	let cursor = null;
	do {
		const data = await graphql(
			`query($owner: String!, $name: String!, $cursor: String) {
				repository(owner: $owner, name: $name) {
					pullRequests(states: OPEN, first: 20, after: $cursor) {
						pageInfo { hasNextPage endCursor }
						nodes { ${PR_FIELDS} }
					}
				}
			}`,
			{ owner, name, cursor },
		);
		const page = data.repository.pullRequests;
		prs.push(...page.nodes);
		cursor = page.pageInfo.hasNextPage ? page.pageInfo.endCursor : null;
	} while (cursor);
	return prs;
}

async function ensureReviewLabel(owner, name) {
	const existing = await api(`/repos/${owner}/${name}/labels/${encodeURIComponent(LABEL_REVIEW)}`, { allow404: true });
	if (existing) return;
	try {
		await api(`/repos/${owner}/${name}/labels`, { method: 'POST', body: REVIEW_LABEL_SPEC });
	} catch (e) {
		if (!String(e.message).includes('already_exists')) throw e;
	}
}

async function applyTurn(owner, name, number, current, decided) {
	if (decided && !current.includes(decided)) {
		await api(`/repos/${owner}/${name}/issues/${number}/labels`, { method: 'POST', body: { labels: [decided] } });
	}
	for (const label of current) {
		if (label !== decided) {
			await api(`/repos/${owner}/${name}/issues/${number}/labels/${encodeURIComponent(label)}`, {
				method: 'DELETE',
				allow404: true,
			});
		}
	}
}

const fmt = (ms) => (Number.isFinite(ms) ? new Date(ms).toISOString() : 'none');

async function main() {
	const repo = process.env.GITHUB_REPOSITORY;
	if (!process.env.GITHUB_TOKEN || !repo) throw new Error('GITHUB_TOKEN and GITHUB_REPOSITORY are required');
	const [owner, name] = repo.split('/');
	const apply = process.env.APPLY === 'true';
	const answeredNow = process.env.ANSWERED_NOW === 'true';
	const prNumber = process.env.PR_NUMBER;

	const prs = prNumber
		? await fetchPr(owner, name, Number(prNumber))
		: await fetchOpenPrs(owner, name);

	if (apply) await ensureReviewLabel(owner, name);

	for (const pr of prs) {
		const current = pr.labels.nodes.map((l) => l.name).filter((l) => l === LABEL_REVIEW || l === LABEL_REPLY);
		const decided = decideTurnLabel(pr, { answeredNow });
		const { ask, answer } = turnTimes(pr, { answeredNow });
		const decidedText = decided ?? `null(${skipReason(pr)})`;
		console.log(
			`#${pr.number} author=${pr.author?.login} current=[${current.join(', ')}] decided=${decidedText}` +
				` ask=${fmt(ask)} answer=${fmt(answer)}`,
		);
		if (apply) await applyTurn(owner, name, pr.number, current, decided);
	}
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
	main().catch((e) => {
		console.error(e);
		process.exit(1);
	});
}
