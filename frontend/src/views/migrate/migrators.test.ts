import {describe, expect, it} from 'vitest'
import {MIGRATORS, isMigratorOfKind} from './migrators'
import type {MigrationKind, MigrationProvider} from '@/client/queries/migration'

const PROVIDERS = Object.keys(MIGRATORS) as MigrationProvider[]
const KINDS: MigrationKind[] = [
	'file',
	'oauth',
	'credentials',
	'csv',
]

describe('MIGRATORS', () => {
	it.each(PROVIDERS)('keys %s by its own id', provider => {
		expect(MIGRATORS[provider].id).toBe(provider)
	})

	it.each(PROVIDERS)('gives %s exactly one kind', provider => {
		expect(KINDS.filter(kind => isMigratorOfKind(provider, kind))).toHaveLength(1)
	})

	it.each([
		['file', ['ticktick', 'vikunja-file', 'wekan']],
		['oauth', ['microsoft-todo', 'todoist', 'trello']],
		['credentials', ['planka']],
		['csv', ['csv']],
	] as const)('derives the %s providers', (kind, providers) => {
		expect(PROVIDERS.filter(provider => isMigratorOfKind(provider, kind)).sort()).toEqual(providers)
	})
})
