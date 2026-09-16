import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

const sdk = vi.hoisted(() => ({sharesList: vi.fn(), sharesCreate: vi.fn(), sharesDelete: vi.fn(), projectsRead: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

import LinkSharing from './LinkSharing.vue'
import {linkShareKeys} from '@/client/queries/linkShares'
import {projectKeys} from '@/client/queries/projects'
import en from '@/i18n/lang/en.json'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

let client: QueryClient
let wrapper: VueWrapper | undefined

beforeEach(() => {
	vi.resetAllMocks()
	vi.useFakeTimers()
	setActivePinia(createPinia())
	client = new QueryClient({defaultOptions: {queries: {retry: false, staleTime: Infinity}}})
	client.setQueryData(linkShareKeys.list(7), [])
	client.setQueryData(projectKeys.detail(7), {id: 7, title: 'Project', views: []})
	sdk.sharesList.mockResolvedValue({data: {items: [], total_pages: 1}})
})

afterEach(() => {
	wrapper?.unmount()
	wrapper = undefined
	vi.useRealTimers()
})

function buttonByText(text: string) {
	const button = wrapper!.findAll('button').find(candidate => candidate.text() === text)
	if (!button) throw new Error(`No button "${text}"`)
	return button
}

describe('LinkSharing', () => {
	it.each([
		['succeeds', () => sdk.sharesCreate.mockResolvedValue({data: {id: 1, hash: 'public-hash'}})],
		['fails', () => sdk.sharesCreate.mockRejectedValue(new Error('nope'))],
	])('keeps no password in the mutation cache after a create %s', async (_outcome, arrange) => {
		arrange()
		wrapper = mount(LinkSharing, {
			props: {projectId: 7},
			global: {
				plugins: [i18n, [VueQueryPlugin, {queryClient: client}]],
				stubs: {
					XButton: {template: '<button type="button" v-bind="$attrs"><slot /></button>'},
					Modal: true,
					Icon: true,
				},
				directives: {tooltip: {}},
			},
		})
		await buttonByText('Create a link share').trigger('click')
		await wrapper.get('#linkShareName').setValue('Client')
		await wrapper.get('#linkSharePassword').setValue('secret')
		await buttonByText('Share').trigger('click')
		await flushPromises()
		await vi.advanceTimersByTimeAsync(0)

		expect(sdk.sharesCreate).toHaveBeenCalledWith(expect.objectContaining({body: expect.objectContaining({password: 'secret'})}))
		expect(client.getMutationCache().getAll().some(mutation => JSON.stringify(mutation.state.variables ?? null).includes('secret'))).toBe(false)
	})
})
