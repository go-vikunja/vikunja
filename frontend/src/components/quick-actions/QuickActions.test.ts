import {describe, it, expect, beforeEach, vi} from 'vitest'
import {defineComponent, h} from 'vue'
import {mount, flushPromises} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'
import {VueQueryPlugin, QueryClient} from '@tanstack/vue-query'

import QuickActions from '@/components/quick-actions/QuickActions.vue'
import {useBaseStore} from '@/stores/base'
import type {IProject} from '@/modelTypes/IProject'
import en from '@/i18n/lang/en.json'

vi.mock('@/services/task', () => ({
	default: class {
		loading = false
		getAll = vi.fn(async () => [])
	},
}))

vi.mock('@/services/team', () => ({
	default: class {
		loading = false
		getAll = vi.fn(async () => [])
	},
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

async function mountQuickActions(project: IProject | null) {
	const errors: unknown[] = []
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{path: '/', name: 'home', component: {template: '<div />'}}],
	})
	await router.push('/')
	await router.isReady()

	// The stores can only be created from within a component setup because the
	// base store itself calls useI18n().
	const Host = defineComponent({
		setup() {
			const baseStore = useBaseStore()
			baseStore.setCurrentProject(project)
			baseStore.setQuickActionsActive(true)
			return () => h(QuickActions)
		},
	})

	const wrapper = mount(Host, {
		global: {
			plugins: [i18n, router, [VueQueryPlugin, {queryClient: new QueryClient()}]],
			stubs: {
				Modal: {template: '<div><slot /></div>'},
				QuickAddMagic: true,
				BaseButton: {template: '<a><slot /></a>'},
				Icon: true,
				XLabel: true,
				SingleTaskInlineReadonly: true,
			},
			directives: {focus: {}},
			config: {
				errorHandler(err: unknown) {
					errors.push(err)
				},
			},
		},
		attachTo: document.body,
	})
	await flushPromises()
	return {wrapper, errors}
}

describe('QuickActions', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('opens without a current project', async () => {
		const {wrapper, errors} = await mountQuickActions(null)

		expect(errors).toEqual([])
		expect(wrapper.find('.quick-actions').exists()).toBe(true)
	})

	it('opens with a current project', async () => {
		const {wrapper, errors} = await mountQuickActions({id: 1, title: 'Test'} as unknown as IProject)

		expect(errors).toEqual([])
		expect(wrapper.find('.quick-actions').exists()).toBe(true)
	})
})
