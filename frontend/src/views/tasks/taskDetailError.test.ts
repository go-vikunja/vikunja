import {describe, expect, it} from 'vitest'

import {taskLoadErrorAction} from './taskDetailError'

describe('taskLoadErrorAction', () => {
	it('ignores the request context fence abort', () => {
		expect(taskLoadErrorAction(new DOMException('Client request context changed', 'AbortError'))).toBe('ignore')
	})

	it('sends 403 and 404 to not found', () => {
		expect(taskLoadErrorAction({status: 403})).toBe('notFound')
		expect(taskLoadErrorAction({status: 404})).toBe('notFound')
	})

	it('reports everything else', () => {
		expect(taskLoadErrorAction({status: 500})).toBe('report')
		expect(taskLoadErrorAction(new Error('boom'))).toBe('report')
	})
})
