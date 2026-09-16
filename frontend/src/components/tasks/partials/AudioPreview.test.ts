import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {nextTick} from 'vue'
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

const mountedPreviews: VueWrapper[] = []

function mountPreview(name = 'memo.mp3') {
	const wrapper = mount(AudioPreview, {
		attachTo: document.body,
		props: {attachment: attachment(name)},
		global: {
			components: {XButton},
			stubs: {Icon: true, RouterLink: true},
			mocks: {$t: (key: string) => key},
		},
	})
	mountedPreviews.push(wrapper)
	return wrapper
}

function unmountPreview(wrapper: VueWrapper) {
	mountedPreviews.splice(mountedPreviews.indexOf(wrapper), 1)
	wrapper.unmount()
}

async function clickPlay(wrapper: VueWrapper) {
	await wrapper.find('button').trigger('click')
}

// The component exposes play() through defineExpose, which the wrapper type does not carry.
function exposedPlay(wrapper: VueWrapper) {
	return (wrapper.vm as unknown as {play: () => Promise<void>}).play()
}

function deferredBlobUrl() {
	let resolveBlobUrl: (url: string) => void = () => {}
	getBlobUrl.mockReturnValue(new Promise<string>(resolve => {
		resolveBlobUrl = resolve
	}))
	return (url: string) => resolveBlobUrl(url)
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
	// The "pause the other player" bookkeeping is module state, so a leftover player leaks into the next test.
	mountedPreviews.splice(0).forEach(wrapper => wrapper.unmount())
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
	})

	it('revokes the object url when unmounted', async () => {
		getBlobUrl.mockResolvedValue('blob:memo')

		const wrapper = mountPreview()
		await clickPlay(wrapper)
		await flushPromises()
		expect(revokeObjectURL).not.toHaveBeenCalled()

		unmountPreview(wrapper)

		expect(revokeObjectURL).toHaveBeenCalledWith('blob:memo')
	})

	it('fetches the file once when the play button is clicked twice in a row', async () => {
		const resolveBlobUrl = deferredBlobUrl()

		const wrapper = mountPreview()
		await clickPlay(wrapper)
		await clickPlay(wrapper)

		expect(getBlobUrl).toHaveBeenCalledTimes(1)

		resolveBlobUrl('blob:memo')
		await flushPromises()

		expect(getBlobUrl).toHaveBeenCalledTimes(1)
		expect(wrapper.find('audio').exists()).toBe(true)
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
	})

	it('brings the play button back and reports when the file cannot be decoded', async () => {
		getBlobUrl.mockResolvedValue('blob:memo')

		const wrapper = mountPreview()
		await clickPlay(wrapper)
		await flushPromises()

		wrapper.find('audio').element.dispatchEvent(new Event('error'))
		await nextTick()

		expect(revokeObjectURL).toHaveBeenCalledWith('blob:memo')
		expect(errorMessage).toHaveBeenCalledWith({message: 'task.attachment.audioError'})
		expect(wrapper.find('audio').exists()).toBe(false)
		expect(wrapper.find('button').exists()).toBe(true)
	})

	it('revokes a blob that arrives after the component was unmounted', async () => {
		const resolveBlobUrl = deferredBlobUrl()

		const wrapper = mountPreview()
		await clickPlay(wrapper)
		unmountPreview(wrapper)

		resolveBlobUrl('blob:memo')
		await flushPromises()

		expect(revokeObjectURL).toHaveBeenCalledTimes(1)
		expect(revokeObjectURL).toHaveBeenCalledWith('blob:memo')
		expect(document.querySelector('audio')).toBeNull()
	})

	it('plays the mounted element when play() is called with the file already loaded', async () => {
		getBlobUrl.mockResolvedValue('blob:memo')

		const wrapper = mountPreview()
		await clickPlay(wrapper)
		await flushPromises()
		const playElement = vi.spyOn(wrapper.find('audio').element, 'play')

		await exposedPlay(wrapper)

		expect(playElement).toHaveBeenCalledTimes(1)
		expect(getBlobUrl).toHaveBeenCalledTimes(1)
	})

	it('does not fetch again when play() is called while the download is in flight', async () => {
		const resolveBlobUrl = deferredBlobUrl()

		const wrapper = mountPreview()
		await clickPlay(wrapper)
		await exposedPlay(wrapper)

		expect(getBlobUrl).toHaveBeenCalledTimes(1)

		resolveBlobUrl('blob:memo')
		await flushPromises()

		expect(getBlobUrl).toHaveBeenCalledTimes(1)
		expect(wrapper.find('audio').exists()).toBe(true)
	})
})
