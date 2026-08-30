import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import TOTP from './TOTP.vue'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'

const get = vi.fn()
const enroll = vi.fn()
const enable = vi.fn()
const disable = vi.fn()
const qrcode = vi.fn(async () => new Blob(['fake-jpeg-bytes']))

vi.mock('@/services/totp', () => ({
	default: class {
		loading = false
		get = get
		enroll = enroll
		enable = enable
		disable = disable
		qrcode = qrcode
	},
}))

vi.mock('@/message', () => ({
	success: vi.fn(),
	error: vi.fn(),
}))

// Avoid the avatar request triggered by setUser.
vi.mock('@/models/user', async (importOriginal) => {
	const original = await importOriginal<typeof import('@/models/user')>()
	return {
		...original,
		fetchAvatarBlobUrl: vi.fn(async () => ''),
		invalidateAvatarCache: vi.fn(),
	}
})

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

let wrapper: VueWrapper | undefined
let errors: unknown[] = []

function mountComponent() {
	return mount(TOTP, {
		global: {
			plugins: [i18n],
			stubs: {
				Card: {template: '<div><slot /></div>'},
				XButton: {
					template: '<button type="button" v-bind="$attrs" @click="$emit(\'click\', $event)"><slot /></button>',
					emits: ['click'],
				},
				FormField: true,
			},
			config: {
				errorHandler(err) {
					errors.push(err)
				},
			},
		},
	})
}

async function mountAndSettle() {
	wrapper = mountComponent()
	await flushPromises()
	return wrapper
}

// Enabled responses omit the secret, so the UI must rely on `enabled` alone.
describe('TOTP settings', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		errors = []
		get.mockReset()
		enroll.mockReset()
		enable.mockReset()
		disable.mockReset()
		qrcode.mockClear()

		const configStore = useConfigStore()
		configStore.totpEnabled = true
		const authStore = useAuthStore()
		authStore.setUser({
			id: 1,
			username: 'user1',
			isLocalUser: true,
		} as never)
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
	})

	it('shows the enroll button when totp is not enrolled', async () => {
		get.mockRejectedValueOnce({response: {data: {code: 1016}}})

		const w = await mountAndSettle()

		expect(w.text()).toContain('Enroll')
		expect(qrcode).not.toHaveBeenCalled()
		expect(errors).toEqual([])
	})

	it('shows the enrollment UI with the qrcode while enrollment is incomplete', async () => {
		get.mockResolvedValueOnce({secret: 'SHAREDSECRET', enabled: false, url: 'otpauth://totp/x'})

		const w = await mountAndSettle()

		expect(w.text()).toContain('SHAREDSECRET')
		expect(qrcode).toHaveBeenCalledTimes(1)
		expect(w.find('img').exists()).toBe(true)
	})

	it('shows the disable UI without the secret or a qrcode request when totp is enabled', async () => {
		get.mockResolvedValueOnce({secret: '', enabled: true, url: ''})

		const w = await mountAndSettle()

		expect(w.text()).toContain("You've successfully set up two factor authentication!")
		expect(w.text()).not.toContain('Enroll')
		expect(w.text()).not.toContain('scan')
		expect(w.find('img').exists()).toBe(false)
		expect(qrcode).not.toHaveBeenCalled()

		const disableBtn = w.findAll('button').find(b => b.text().toLowerCase().includes('disable'))
		expect(disableBtn).toBeTruthy()
		await disableBtn!.trigger('click')
		await flushPromises()
		expect(qrcode).not.toHaveBeenCalled()
		expect(errors).toEqual([])
	})
})
