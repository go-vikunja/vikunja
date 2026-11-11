import {mount} from '@vue/test-utils'
import {createMemoryHistory, createRouter, RouterView} from 'vue-router'
import {h} from 'vue'
import {expect, it, vi} from 'vitest'
import {GLOBAL_NOW_INTERVAL, useGlobalNow} from './useGlobalNow'

it('can start the shared clock during rendering and keeps it ticking', async () => {
	vi.useFakeTimers()
	const warnings: string[] = []
	const clock = {render: () => h('time', useGlobalNow().now.value.toISOString())}
	const router = createRouter({history: createMemoryHistory(), routes: [{path: '/', component: clock}]})
	await router.push('/')
	await router.isReady()
	const wrapper = mount(RouterView, {
		global: {plugins: [router], config: {warnHandler: message => warnings.push(message)}},
	})
	try {
		expect(warnings).toEqual([])
		const before = wrapper.text()
		await vi.advanceTimersByTimeAsync(GLOBAL_NOW_INTERVAL)
		expect(wrapper.text()).not.toBe(before)
	} finally {
		wrapper.unmount()
		vi.clearAllTimers()
		vi.useRealTimers()
	}
})
