import {describe, it, expectTypeOf} from 'vitest'
import {useGanttTaskList, type UseGanttTaskListReturn} from './useGanttTaskList'
import type {Filters} from '@/composables/useRouteFilters'

interface TestFilters extends Filters {
	projectId: number
}

describe('useGanttTaskList return type', () => {
	it('should match interface', () => {
		type Result = ReturnType<typeof useGanttTaskList<TestFilters>>
		expectTypeOf<Result>().toEqualTypeOf<UseGanttTaskListReturn>()
	})
})
