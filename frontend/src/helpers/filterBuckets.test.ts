import {describe, expect, it} from 'vitest'
import {
	matchesBucketFilter,
	parseBucketFilter,
	planBucketFilterMove,
	type BucketFilterField,
} from '@/helpers/filterBuckets'

const bucket = (filter: string, s = '') => parseBucketFilter({
	filter,
	s,
})

function task(usernames: string[] = [], labelIds: number[] = []) {
	return {
		assignees: usernames.map((username, index) => ({
			id: index + 1,
			username,
		})),
		labels: labelIds.map(id => ({
			id,
			title: `label ${id}`,
		})),
	}
}

describe('parseBucketFilter', () => {
	it.each([
		'assignees in alice',
		'assignees = alice',
		'assignees not in alice, bob, carol',
		'assignees != alice, bob',
		'assignees ?!= alice, bob',
		'labels ?= 4',
		'labels in 5',
		'labels in "5"',
		'labels in 5 && assignees in alice',
	])('accepts %s', filter => {
		expect(bucket(filter)).not.toBeNull()
	})

	it.each([
		'',
		'assignees in alice, bob',
		'assignees = alice, bob',
		'labels in 5 || labels in 6',
		'(labels in 5) && assignees in alice',
		'labels in 5 && priority > 3',
		'due_date < now',
		'done = false',
		'assignees like jak',
		'Assignees IN alice',
		'labels in "foo bar"',
	])('rejects %s', filter => {
		expect(bucket(filter)).toBeNull()
	})

	it('rejects a bucket that also searches', () => {
		expect(bucket('assignees in alice', 'foo')).toBeNull()
	})

	it('strips quotes from values', () => {
		expect(bucket('assignees in "alice"')).toEqual([{
			field: 'assignees',
			negated: false,
			values: ['alice'],
		}])
	})

	it('unquotes a whole list before splitting it', () => {
		expect(bucket('assignees not in "alice, bob"')?.[0].values).toEqual(['alice', 'bob'])
	})

	it('unquotes each value of a list quoted per item', () => {
		expect(bucket('assignees not in "alice", "bob"')?.[0].values).toEqual(['alice', 'bob'])
	})
})

describe('matchesBucketFilter', () => {
	it('requires every clause', () => {
		const filter = bucket('labels in 5 && assignees in alice')!
		expect(matchesBucketFilter(task(['alice'], [5]), filter)).toBe(true)
		expect(matchesBucketFilter(task(['alice']), filter)).toBe(false)
	})

	it('matches labels by numeric id and users by exact username', () => {
		expect(matchesBucketFilter(task([], [5]), bucket('labels in 05')!)).toBe(true)
		expect(matchesBucketFilter(task(['Alice']), bucket('assignees in alice')!)).toBe(false)
	})

	it('matches the catch-all when none of the excluded values are present', () => {
		const filter = bucket('assignees not in alice, bob')!
		expect(matchesBucketFilter(task(), filter)).toBe(true)
		expect(matchesBucketFilter(task(['dave']), filter)).toBe(true)
		expect(matchesBucketFilter(task(['bob']), filter)).toBe(false)
	})

	it('matches tasks without a value when nulls are included', () => {
		const filter = bucket('assignees in alice')!
		expect(matchesBucketFilter(task(), filter, true)).toBe(true)
		expect(matchesBucketFilter(task(['bob']), filter, true)).toBe(false)
	})
})

describe('planBucketFilterMove', () => {
	const edit = (field: BucketFilterField, value: string, add: boolean) => ({
		field,
		value,
		add,
	})
	const add = (field: BucketFilterField, value: string) => edit(field, value, true)
	const remove = (field: BucketFilterField, value: string) => edit(field, value, false)

	it('reassigns between single-assignee buckets', () => {
		const edits = planBucketFilterMove(
			task(['alice']),
			bucket('assignees in alice'),
			bucket('assignees in bob')!,
		)
		expect(edits).toEqual([
			add('assignees', 'bob'),
			remove('assignees', 'alice'),
		])
	})

	it('assigns from the catch-all without removing anything', () => {
		const edits = planBucketFilterMove(
			task(),
			bucket('assignees not in alice, bob'),
			bucket('assignees in alice')!,
		)
		expect(edits).toEqual([
			add('assignees', 'alice'),
		])
	})

	it('unassigns every excluded user when dropped on the catch-all', () => {
		const edits = planBucketFilterMove(
			task(['alice', 'bob', 'dave']),
			bucket('assignees in alice'),
			bucket('assignees not in alice, bob')!,
		)
		expect(edits).toEqual([
			remove('assignees', 'alice'),
			remove('assignees', 'bob'),
		])
	})

	it('applies every value of a multi-clause target in one drop', () => {
		const edits = planBucketFilterMove(
			task(),
			bucket('assignees not in alice'),
			bucket('labels in 5 && assignees in alice')!,
		)
		expect(edits).toEqual([
			add('labels', '5'),
			add('assignees', 'alice'),
		])
	})

	it('keeps source values the target also requires', () => {
		const edits = planBucketFilterMove(
			task(['alice'], [5]),
			bucket('labels in 5 && assignees in alice'),
			bucket('labels in 6 && assignees in alice')!,
		)
		expect(edits).toEqual([
			add('labels', '6'),
			remove('labels', '5'),
		])
	})

	it('only applies the target when the source bucket is not editable', () => {
		expect(planBucketFilterMove(task(['bob']), null, bucket('assignees in alice')!)).toEqual([
			add('assignees', 'alice'),
		])
	})

	it('removes a value only once when the source requires it and the target excludes it', () => {
		const edits = planBucketFilterMove(
			task(['alice']),
			bucket('assignees in alice'),
			bucket('assignees not in alice, bob')!,
		)
		expect(edits).toEqual([
			remove('assignees', 'alice'),
		])
	})

	it('skips values the task already has', () => {
		const edits = planBucketFilterMove(task(['alice'], [6]), bucket('assignees in alice'), bucket('labels in 6')!)
		expect(edits).toEqual([
			remove('assignees', 'alice'),
		])
	})
})
