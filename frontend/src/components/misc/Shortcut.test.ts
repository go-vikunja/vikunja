import {describe, it, expect} from 'vitest'
import {mount} from '@vue/test-utils'

import Shortcut from './Shortcut.vue'

describe('Shortcut.vue', () => {
	it('renders a regular key combination with the default separator', () => {
		const wrapper = mount(Shortcut, {
			props: {
				keys: ['ctrl', 'e'],
			},
		})

		expect(wrapper.findAll('kbd').map(kbd => kbd.text())).toEqual(['ctrl', 'e'])
		expect(wrapper.text()).toContain('ctrl+e')
	})

	it('renders a key sequence with a custom separator', () => {
		const wrapper = mount(Shortcut, {
			props: {
				keys: ['g', 'o'],
				combination: 'then',
			},
		})

		expect(wrapper.findAll('kbd').map(kbd => kbd.text())).toEqual(['g', 'o'])
		expect(wrapper.text()).toContain('gtheno')
	})

	it('renders special characters as visible shortcut keys', () => {
		const wrapper = mount(Shortcut, {
			props: {
				keys: ['shift', '/'],
			},
		})

		expect(wrapper.findAll('kbd').map(kbd => kbd.text())).toEqual(['shift', '/'])
		expect(wrapper.text()).toContain('shift+/')
	})
})
