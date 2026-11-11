import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {beforeEach, describe, it, expect, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import LinkedAccounts from './LinkedAccounts.vue'
import Card from '@/components/misc/Card.vue'
import en from '@/i18n/lang/en.json'

const {sdk} = vi.hoisted(() => ({
	sdk: {
		totpGet: vi.fn(),
		userIdentitiesList: vi.fn(),
		userIdentitiesLink: vi.fn(),
		userIdentitiesUnlink: vi.fn(),
	},
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', async importOriginal => ({
	...await importOriginal<typeof import('@/message')>(),
	success: vi.fn(),
	error: vi.fn(),
}))
vi.mock('@/helpers/time/formatDate', () => ({formatDisplayDate: () => ''}))
vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({
		totp_enabled: true,
		auth: {
			openid_connect: {
				enabled: true,
				providers: [],
			},
		},
	}),
}))

const i18n = createI18n({
	legacy: false,
	locale: 'en',
	messages: {en},
})

const queryClientConfig = {
	defaultOptions: {
		queries: {retry: false},
	},
}

async function mountWithUnlinkModal() {
	const wrapper = mount(LinkedAccounts, {
		global: {
			plugins: [i18n, [VueQueryPlugin, {queryClient: new QueryClient(queryClientConfig)}]],
			components: {Card},
			stubs: {
				XButton: true,
				Modal: {template: '<div><slot /></div>'},
			},
		},
	})
	await flushPromises()
	await wrapper.find('x-button-stub[variant="secondary"]').trigger('click')
	return wrapper
}

beforeEach(() => {
	vi.clearAllMocks()
	sdk.userIdentitiesList.mockResolvedValue({
		data: {
			items: [{
				id: 1,
				provider_key: 'dex',
				provider_name: 'Dex',
			}],
		},
	})
	sdk.userIdentitiesUnlink.mockResolvedValue({})
})

describe('LinkedAccounts unlink', () => {
	it('asks for the TOTP passcode when the TOTP status cannot be loaded', async () => {
		sdk.totpGet.mockRejectedValue({status: 500})
		const wrapper = await mountWithUnlinkModal()
		await flushPromises()

		expect(wrapper.find('#unlinkIdentityPasscode').exists()).toBe(true)
	})

	it('does not unlink while the TOTP status is loading', async () => {
		sdk.totpGet.mockReturnValue(new Promise(() => {}))
		const wrapper = await mountWithUnlinkModal()
		const password = wrapper.find('#unlinkIdentityPassword')
		await password.setValue('secret')
		await password.trigger('keyup.enter')
		await flushPromises()

		expect(sdk.userIdentitiesUnlink).not.toHaveBeenCalled()
	})

	it('hides the TOTP passcode when TOTP is not enrolled', async () => {
		sdk.totpGet.mockResolvedValue({data: {enabled: false}})
		const wrapper = await mountWithUnlinkModal()
		await flushPromises()

		expect(wrapper.find('#unlinkIdentityPasscode').exists()).toBe(false)
	})
})
