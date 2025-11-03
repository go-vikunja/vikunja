import {describe, it, expectTypeOf} from 'vitest'
import {useRouteFilters, type UseRouteFiltersReturn} from './useRouteFilters'

interface DummyFilters {foo: string}

describe('useRouteFilters type inference', () => {
	it('should infer return types based on filter interface', () => {
		type Result = ReturnType<typeof useRouteFilters<DummyFilters>>
		expectTypeOf<Result>().toEqualTypeOf<UseRouteFiltersReturn<DummyFilters>>()
	})
})
