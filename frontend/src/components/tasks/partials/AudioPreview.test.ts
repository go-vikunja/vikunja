import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import AudioPreview from './AudioPreview.vue'
import XButton from '@/components/input/Button.vue'
import type {IAttachment} from '@/modelTypes/IAttachment'

const {getBlobUrl} = vi.hoisted(() => ({getBlobUrl: vi.fn()}))
const {errorMessage} = vi.hoisted(() => ({errorMessage: vi.fn()}))

vi.mock('@/services/attachment', () => ({
	default: class {
		getBlobUrl = getBlobUrl
	},
}))

vi.mock('@/message', () => ({error: errorMessage}))

vi.mock('vue-i18n', () => ({useI18n: () => ({t: (key: string) => key})}))

function attachment(name: string): IAttachment {
	return {id: 1, taskId: 1, file: {name, mime: 'audio/mpeg'}} as unknown as IAttachment
}

function mountPreview(name = 'memo.mp3') {
	return mount(AudioPreview, {
		attachTo: document.body,
		props: {attachment: attachment(name)},
		global: {
			components: {XButton},
			stubs: {Icon: true, RouterLink: true},
			mocks: {$t: (key: string) => key},
		},
	})
}

async function clickPlay(wrapper: VueWrapper) {
	await wrapper.find('button').trigger('click')
}

let revokeObjectURL: ReturnType<typeof vi.fn<(url: string) => void>>

beforeEach(() => {
	// happy-dom lacks media methods; plain functions, not vi.fn: a shared prototype mock merges every element's calls.
	HTMLMediaElement.prototype.play = () => Promise.resolve()
	HTMLMediaElement.prototype.pause = () => {}
	revokeObjectURL = vi.fn<(url: string) => void>()
	window.URL.revokeObjectURL = revokeObjectURL
})

afterEach(() => {
	getBlobUrl.mockReset()
	errorMessage.mockReset()
	document.body.innerHTML = ''
})

describe('AudioPreview.vue', () => {
	it('pauses the player that was running when another one starts', async () => {
		getBlobUrl.mockResolvedValueOnce('blob:a').mockResolvedValueOnce('blob:b')

		const first = mountPreview('a.mp3')
		const second = mountPreview('b.mp3')
		await clickPlay(first)
		await clickPlay(second)
		await flushPromises()

		const firstPlayer = first.find('audio').element
		const secondPlayer = second.find('audio').element
		const pauseFirst = vi.spyOn(firstPlayer, 'pause')
		const pauseSecond = vi.spyOn(secondPlayer, 'pause')

		firstPlayer.dispatchEvent(new Event('play'))
		secondPlayer.dispatchEvent(new Event('play'))

		expect(pauseFirst).toHaveBeenCalledTimes(1)
		expect(pauseSecond).not.toHaveBeenCalled()

		first.unmount()
		second.unmount()
	})

	it('revokes the object url when unmounted', async () => {
		getBlobUrl.mockResolvedValue('blob:memo')

		const wrapper = mountPreview()
		await clickPlay(wrapper)
		await flushPromises()
		expect(revokeObjectURL).not.toHaveBeenCalled()

		wrapper.unmount()

		expect(revokeObjectURL).toHaveBeenCalledWith('blob:memo')
	})

	it('fetches the file once when the play button is clicked twice in a row', async () => {
		let resolveBlobUrl: (url: string) => void = () => {}
		getBlobUrl.mockReturnValue(new Promise(resolve => {
			resolveBlobUrl = resolve
		}))

		const wrapper = mountPreview()
		await clickPlay(wrapper)
		await clickPlay(wrapper)

		expect(getBlobUrl).toHaveBeenCalledTimes(1)

		resolveBlobUrl('blob:memo')
		await flushPromises()

		expect(getBlobUrl).toHaveBeenCalledTimes(1)
		expect(wrapper.find('audio').exists()).toBe(true)

		wrapper.unmount()
	})

	it('surfaces a failed download and leaves the play button usable', async () => {
		const downloadFailed = new Error('nope')
		getBlobUrl.mockRejectedValueOnce(downloadFailed).mockResolvedValueOnce('blob:memo')

		const wrapper = mountPreview()
		await clickPlay(wrapper)
		await flushPromises()

		expect(errorMessage).toHaveBeenCalledWith(downloadFailed)
		expect(wrapper.find('audio').exists()).toBe(false)
		expect(wrapper.find('button').attributes('disabled')).toBeUndefined()

		await clickPlay(wrapper)
		await flushPromises()

		expect(getBlobUrl).toHaveBeenCalledTimes(2)
		expect(wrapper.find('audio').exists()).toBe(true)

		wrapper.unmount()
	})
})
