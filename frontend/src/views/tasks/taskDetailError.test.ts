import {describe, expect, it} from 'vitest'

import {ApiError} from '@/client/problem'
import {taskLoadErrorAction} from './taskDetailError'

describe('taskLoadErrorAction', () => {
	it('ignores the request context fence abort', () => {
		const abort = new DOMException('Client request context changed', 'AbortError')
		expect(taskLoadErrorAction(abort)).toEqual({action: 'ignore'})
	})

	it('keeps trashed tasks in place with their trash details', () => {
		const inTrash = {
			project_id: 1,
			deleted_at: '2026-01-01T00:00:00Z',
		}
		expect(taskLoadErrorAction(new ApiError({
			status: 404,
			code: 4040,
			detail: 'in trash',
			details: inTrash,
		}))).toEqual({
			action: 'inTrash',
			inTrash,
		})
	})

	it('sends 403 and 404 to not found', () => {
		expect(taskLoadErrorAction({status: 403})).toEqual({action: 'notFound'})
		expect(taskLoadErrorAction({status: 404})).toEqual({action: 'notFound'})
	})

	it('reports everything else', () => {
		expect(taskLoadErrorAction({status: 500})).toEqual({action: 'report'})
		expect(taskLoadErrorAction(new Error('boom'))).toEqual({action: 'report'})
	})
})
