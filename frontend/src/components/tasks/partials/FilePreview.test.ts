import {
	describe,
	it,
	expect,
	beforeEach,
	afterEach,
	vi,
} from 'vitest'
import {
	mount,
	flushPromises,
	type VueWrapper,
} from '@vue/test-utils'
import FilePreview from './FilePreview.vue'
import {queryClient} from '@/client/queryClient'
import {attachmentKeys} from '@/client/queries/attachments'
import type {TaskAttachment as IAttachment} from '@/client/generated'

const {getBlobUrl} = vi.hoisted(() => ({getBlobUrl: vi.fn()}))

vi.mock('@/client/generated', () => ({taskAttachmentsDownload: getBlobUrl}))

function attachment(): IAttachment {
	return {
		id: 1,
		task_id: 1,
		file: {
			name: 'chart.png',
			mime: 'image/png',
		},
	} as unknown as IAttachment
}

const mountedPreviews: VueWrapper[] = []

function mountPreview() {
	const wrapper = mount(FilePreview, {
		props: {modelValue: attachment()},
		global: {stubs: {Icon: true}},
	})
	mountedPreviews.push(wrapper)
	return wrapper
}

beforeEach(() => {
	queryClient.removeQueries({queryKey: attachmentKeys.blobs})
	URL.createObjectURL = vi.fn(blob => (blob as Blob & {testUrl?: string}).testUrl ?? 'blob:real-attachment')
	window.URL.revokeObjectURL = vi.fn()
	getBlobUrl.mockReset()
})

afterEach(() => {
	mountedPreviews.splice(0).forEach(wrapper => wrapper.unmount())
})

describe('FilePreview.vue', () => {
	it('keeps the thumbnail when a list refetch hands over a new object with the same ids', async () => {
		getBlobUrl.mockResolvedValue({data: Object.assign(new Blob(['bytes']), {testUrl: 'blob:md'})})

		const wrapper = mountPreview()
		await flushPromises()
		expect(wrapper.find('img').attributes('src')).toBe('blob:md')

		await wrapper.setProps({modelValue: attachment()})

		expect(wrapper.find('img').attributes('src')).toBe('blob:md')
		expect(getBlobUrl).toHaveBeenCalledTimes(1)
	})

	it('requests the md preview and revokes the url it owns when it unmounts', async () => {
		getBlobUrl.mockResolvedValue({data: Object.assign(new Blob(['bytes']), {testUrl: 'blob:md'})})

		const wrapper = mountPreview()
		await flushPromises()

		expect(getBlobUrl).toHaveBeenCalledWith(expect.objectContaining({
			path: {
				task: 1,
				attachment: 1,
			},
			query: {preview_size: 'md'},
		}))
		expect(window.URL.revokeObjectURL).not.toHaveBeenCalled()

		wrapper.unmount()
		mountedPreviews.length = 0

		expect(window.URL.revokeObjectURL).toHaveBeenCalledExactlyOnceWith('blob:md')
	})

	it('hands every preview of the same attachment its own url over one request', async () => {
		getBlobUrl.mockResolvedValue({data: Object.assign(new Blob(['bytes']), {testUrl: 'blob:md'})})

		mountPreview()
		mountPreview()
		await flushPromises()

		expect(getBlobUrl).toHaveBeenCalledTimes(1)
		expect(URL.createObjectURL).toHaveBeenCalledTimes(2)
	})
})
