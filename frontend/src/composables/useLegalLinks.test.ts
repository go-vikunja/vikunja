import {describe, it, expect, beforeEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {defineComponent, h} from 'vue'
import {mount} from '@vue/test-utils'

import {useLegalLinks} from './useLegalLinks'
import {useConfigStore} from '@/stores/config'

function runLegalLinks() {
	let result: ReturnType<typeof useLegalLinks> | undefined
	const Comp = defineComponent({
		setup() {
			result = useLegalLinks()
			return () => h('div')
		},
	})
	mount(Comp, {global: {plugins: [createI18n({
		legacy: false,
		locale: 'de',
		messages: {de: {navigation: {imprint: 'Impressum', privacy: 'Datenschutz'}}},
	})]}})
	return result!.value
}

describe('useLegalLinks', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('translates the well-known texts and keeps the rest as-is, inherited object keys included', () => {
		useConfigStore().setConfig({
			version: 'v2',
			legal: {
				links: [
					{
						text: 'Imprint',
						url: 'https://example.test/imprint',
					},
					{
						text: 'Privacy Policy',
						url: 'https://example.test/privacy',
					},
					{
						text: 'Terms',
						url: 'https://example.test/terms',
					},
					{
						text: 'constructor',
						url: 'https://example.test/constructor',
					},
				],
			},
		})

		expect(runLegalLinks()).toEqual([
			{
				text: 'Impressum',
				url: 'https://example.test/imprint',
			},
			{
				text: 'Datenschutz',
				url: 'https://example.test/privacy',
			},
			{
				text: 'Terms',
				url: 'https://example.test/terms',
			},
			{
				text: 'constructor',
				url: 'https://example.test/constructor',
			},
		])
	})
})
