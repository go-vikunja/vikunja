import { test } from 'node:test';
import assert from 'node:assert/strict';
import { decideTurnLabel } from './review-turn.mjs';

const AUTHOR = 'outsider';
const MAINT = 'kolaente';
const T = {
	t1: '2026-01-01T10:00:00Z',
	t2: '2026-01-02T10:00:00Z',
	t3: '2026-01-03T10:00:00Z',
};
const REVIEW = 'waiting for review';
const REPLY = 'waiting for reply';

const pr = (nodes, overrides = {}) => ({
	number: 1,
	state: 'OPEN',
	isDraft: false,
	authorAssociation: 'FIRST_TIME_CONTRIBUTOR',
	author: { login: AUTHOR, __typename: 'User' },
	labels: { nodes: [] },
	timelineItems: { nodes },
	...overrides,
});

const changesRequested = (at, login = MAINT, association = 'MEMBER') => ({
	__typename: 'PullRequestReview',
	submittedAt: at,
	state: 'CHANGES_REQUESTED',
	authorAssociation: association,
	author: { login },
});
const comment = (at, login = AUTHOR) => ({ __typename: 'IssueComment', createdAt: at, author: { login } });
const reviewBy = (at, login = AUTHOR) => ({
	__typename: 'PullRequestReview',
	submittedAt: at,
	state: 'COMMENTED',
	authorAssociation: 'NONE',
	author: { login },
});
const commit = (at) => ({ __typename: 'PullRequestCommit', commit: { committedDate: at } });
const forcePush = (at) => ({ __typename: 'HeadRefForcePushedEvent', createdAt: at });
const readyForReview = (at) => ({ __typename: 'ReadyForReviewEvent', createdAt: at });
const labeled = (at, actorType = 'User') => ({
	__typename: 'LabeledEvent',
	createdAt: at,
	label: { name: REPLY },
	actor: { __typename: actorType },
});
const unlabeled = (at) => ({ __typename: 'UnlabeledEvent', createdAt: at, label: { name: REPLY } });

test('never reviewed is waiting for review', () => {
	assert.equal(decideTurnLabel(pr([])), REVIEW);
});

test('changes requested with no author activity is waiting for reply', () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1)])), REPLY);
});

test('author comment after changes requested is waiting for review', () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1), comment(T.t2)])), REVIEW);
});

test('author review-thread reply after changes requested is waiting for review', () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1), reviewBy(T.t2)])), REVIEW);
});

test('commit after changes requested is waiting for review', () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1), commit(T.t2)])), REVIEW);
});

test('force push after changes requested is waiting for review', () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1), forcePush(T.t2)])), REVIEW);
});

test('ready for review after changes requested is waiting for review', () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1), readyForReview(T.t2)])), REVIEW);
});

test('author activity before changes requested does not count as an answer', () => {
	assert.equal(decideTurnLabel(pr([comment(T.t1), changesRequested(T.t2)])), REPLY);
});

test('hand-applied waiting for reply label with no review is waiting for reply', () => {
	assert.equal(decideTurnLabel(pr([labeled(T.t1)])), REPLY);
});

test('hand-applied waiting for reply label removed afterwards is waiting for review', () => {
	assert.equal(decideTurnLabel(pr([labeled(T.t1), unlabeled(T.t2)])), REVIEW);
});

test('bot-applied waiting for reply label alone is not an ask', () => {
	assert.equal(decideTurnLabel(pr([labeled(T.t1, 'Bot')])), REVIEW);
});

test("a non-maintainer's changes requested review does not count", () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1, 'stranger', 'NONE')])), REVIEW);
});

test('draft is out of scope', () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1)], { isDraft: true })), null);
});

test('maintainer-authored PR is out of scope', () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1)], { authorAssociation: 'MEMBER' })), null);
});

test('bot-authored PR is out of scope', () => {
	assert.equal(
		decideTurnLabel(pr([], { author: { login: 'renovate', __typename: 'Bot' } })),
		null,
	);
});

test('a maintainer reviewing their own PR does not count as an ask', () => {
	assert.equal(
		decideTurnLabel(pr([changesRequested(T.t1, AUTHOR, 'MEMBER')])),
		REVIEW,
	);
});

test('answeredNow turns an outstanding ask into waiting for review', () => {
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1)])), REPLY);
	assert.equal(decideTurnLabel(pr([changesRequested(T.t1)]), { answeredNow: true }), REVIEW);
});

test('closed PR is out of scope', () => {
	assert.equal(decideTurnLabel(pr([], { state: 'CLOSED' })), null);
});
