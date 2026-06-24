import {describe, it, expect} from 'vitest'

import {objectToSnakeCase} from '@/helpers/case'

// Regression test: objectToSnakeCase mangles hyphenated permission group
// names like "time-entries" → "time_entries". ApiTokenService.beforeCreate
// works around this by restoring the original permissions after the transform.
// This test ensures the underlying problem is documented and will tell us
// if the library behaviour changes.
describe('objectToSnakeCase on API token permissions', () => {
	it('mangles time-entries to time_entries (the bug we work around)', () => {
		const input = {
			title: 'test',
			expiresAt: '2099-01-01T00:00:00Z',
			permissions: {
				'tasks': ['read_all', 'create'],
				'time-entries': ['read_all', 'create'],
				'tasks_assignees': ['create', 'delete'],
			},
		}

		const result = objectToSnakeCase(input)

		// The outer key is fine
		expect(result.expires_at).toBe('2099-01-01T00:00:00Z')

		// time-entries gets mangled — this is the bug
		expect(result.permissions['time_entries']).toEqual(['read_all', 'create'])
		expect(result.permissions['time-entries']).toBeUndefined()

		// Other groups survive
		expect(result.permissions['tasks']).toEqual(['read_all', 'create'])
		expect(result.permissions['tasks_assignees']).toEqual(['create', 'delete'])
	})
})
