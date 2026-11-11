import {describe, expect, it} from 'vitest'
import type {ProjectView} from '@/client/generated'

import {getDefaultProjectView} from './projectView'

const views = [
	{
		id: 1,
		view_kind: 'list',
	},
	{
		id: 2,
		view_kind: 'kanban',
	},
] satisfies ProjectView[]

describe('getDefaultProjectView', () => {
	it('uses the first view when configured', () => {
		expect(getDefaultProjectView(views, 'first')).toBe(views[0])
	})

	it('selects the preferred view kind', () => {
		expect(getDefaultProjectView(views, 'kanban')).toBe(views[1])
	})

	it('falls back to the first view when the preferred kind is missing', () => {
		expect(getDefaultProjectView(views, 'table')).toBe(views[0])
	})

	it('returns no view while the project has no views', () => {
		expect(getDefaultProjectView([], 'first')).toBeUndefined()
	})
})
