import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {VueQueryPlugin} from '@tanstack/vue-query'

import MigrationCSV from './MigrationCSV.vue'
import {queryClient} from '@/client/queryClient'
import en from '@/i18n/lang/en.json'

const sdk = vi.hoisted(() => ({
	migrationCsvDetect: vi.fn(),
	migrationCsvPreview: vi.fn(),
}))

vi.mock('@/client/generated', async importOriginal => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	...sdk,
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

let wrapper: VueWrapper | undefined

describe('MigrationCSV', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		queryClient.clear()
		sdk.migrationCsvDetect.mockReset().mockResolvedValue({
			data: {
				columns: ['Title', 'Tags'],
				delimiter: ',',
				quote_char: '"',
				date_format: '2006-01-02',
				suggested_mapping: [
					{column_index: 0, column_name: 'Title', attribute: 'title'},
					{column_index: 1, column_name: 'Tags', attribute: 'labels'},
				],
				preview_rows: [['Task', 'a,b']],
			},
		})
		sdk.migrationCsvPreview.mockReset()
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
	})

	it('shows a rejected preview and clears it once a preview succeeds', async () => {
		sdk.migrationCsvPreview
			.mockRejectedValueOnce({status: 400, code: 14042, i18n_params: {row: '3', max: '100'}, detail: 'too many labels'})
			.mockResolvedValueOnce({data: {tasks: [], total_rows: 1}})

		wrapper = mount(MigrationCSV, {
			global: {
				plugins: [i18n, [VueQueryPlugin, {queryClient}]],
				stubs: {SingleTaskInProject: true},
			},
		})

		const input = wrapper.find('input[type="file"]')
		Object.defineProperty(input.element, 'files', {value: [new File(['Title,Tags'], 'import.csv')]})
		await input.trigger('change')
		await flushPromises()

		expect(wrapper.text()).toContain('Row 3 of the CSV file has more than 100 labels.')

		await wrapper.findAll('select').at(-1)!.setValue('ignore')
		await flushPromises()

		expect(sdk.migrationCsvPreview).toHaveBeenCalledTimes(2)
		expect(wrapper.text()).not.toContain('more than 100 labels')
	})
})
