import {describe, it, expect} from 'vitest'
import {mount} from '@vue/test-utils'
import PaginationItem from './PaginationItem.vue'

const mountItem = (props: Record<string, unknown>) => mount(PaginationItem, {
	props: {
		variant: 'previous',
		...props,
	},
	slots: {
		default: 'Previous',
	},
	global: {
		stubs: {
			RouterLink: {
				template: '<a :href="typeof to === \'string\' ? to : \'#\'"><slot /></a>',
				props: ['to'],
			},
		},
	},
})

describe('PaginationItem', () => {
	it('renders a link when enabled', () => {
		const wrapper = mountItem({to: '/projects/1/1?page=2'})
		expect(wrapper.find('a').exists()).toBe(true)
	})

	it('does not render a navigable link when disabled', () => {
		const wrapper = mountItem({to: '/projects/1/1?page=0', disabled: true})
		expect(wrapper.find('a').exists()).toBe(false)
	})
})
