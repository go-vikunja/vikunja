import {describe, it, expect, beforeEach, afterEach} from 'vitest'
import {defineComponent, h, nextTick, type Ref} from 'vue'
import {mount, flushPromises, enableAutoUnmount} from '@vue/test-utils'
import {createRouter, createMemoryHistory, type Router} from 'vue-router'

import {useBackdropRouteQuery} from './useBackdropRouteQuery'

type QueryState = {
	query?: Ref<string | undefined>,
}

async function mountQuery(path: string, backdropView?: string): Promise<{router: Router, state: QueryState}> {
	window.history.replaceState(backdropView ? {backdropView} : {}, '')

	const router = createRouter({
		history: createMemoryHistory(),
		routes: [
			{path: '/projects/:projectId/:viewId', name: 'project.view', component: {render: () => null}},
			{path: '/tasks/:id', name: 'task.detail', component: {render: () => null}},
		],
	})
	await router.push(path)
	await router.isReady()

	const state: QueryState = {}
	const TestComponent = defineComponent({
		setup() {
			state.query = useBackdropRouteQuery('filter')
			return () => h('div')
		},
	})

	mount(TestComponent, {global: {plugins: [router]}})
	await nextTick()
	return {router, state}
}

function queryFrom(state: QueryState): Ref<string | undefined> {
	if (!state.query) {
		throw new Error('query was not initialized')
	}
	return state.query
}

enableAutoUnmount(afterEach)

beforeEach(() => {
	window.history.replaceState({}, '')
})

describe('useBackdropRouteQuery', () => {
	it('reads and writes the current route query without a backdrop', async () => {
		const {router, state} = await mountQuery('/projects/1/4?filter=done%3Dtrue')
		const query = queryFrom(state)

		expect(query.value).toBe('done=true')

		query.value = 'done=false'
		await flushPromises()
		await nextTick()

		expect(router.currentRoute.value.query.filter).toBe('done=false')
	})

	it('reads the backdrop query while the task modal is open', async () => {
		const {router, state} = await mountQuery('/projects/1/4?filter=done%3Dtrue')
		const query = queryFrom(state)

		window.history.replaceState({backdropView: '/projects/1/4?filter=done%3Dtrue'}, '')
		await router.push('/tasks/1')
		await nextTick()

		expect(query.value).toBe('done=true')

		// Writes are ignored while the background view cannot be interacted with.
		query.value = 'done=false'
		await flushPromises()
		await nextTick()

		expect(router.currentRoute.value.fullPath).toBe('/tasks/1')

		window.history.replaceState({}, '')
		await router.push('/projects/1/4?filter=done%3Dtrue')
		await nextTick()

		expect(query.value).toBe('done=true')
	})

	it('reads the backdrop query when mounted behind a modal', async () => {
		const {state} = await mountQuery('/tasks/1', '/projects/1/4?filter=done%3Dtrue')

		expect(queryFrom(state).value).toBe('done=true')
	})
})
