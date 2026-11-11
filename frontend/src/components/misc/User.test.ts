import {mount} from '@vue/test-utils'
import {expect, it, vi} from 'vitest'
import {createI18n} from 'vue-i18n'

import User from './User.vue'
import en from '@/i18n/lang/en.json'

vi.mock('@/components/misc/UserAvatar.vue', () => ({default: {template: '<span />'}}))

const i18n = createI18n({
	legacy: false,
	locale: 'en',
	messages: {en},
})

function mountUser(user: InstanceType<typeof User>['$props']['user']) {
	return mount(User, {
		props: {user},
		global: {
			plugins: [i18n],
			directives: {tooltip: {}},
		},
	})
}

it('names the owner on the bot badge', () => {
	const wrapper = mountUser({
		username: 'bot-helper',
		bot_owner_id: 21,
		bot_owner: {name: 'Owner A'},
	})
	expect(wrapper.find('.bot-badge').attributes('aria-label')).toBe('Bot, created by Owner A')
})

it('shows a plain bot badge without an owner', () => {
	const wrapper = mountUser({username: 'bot-helper', bot_owner_id: 21})
	expect(wrapper.find('.bot-badge').attributes('aria-label')).toBe('Bot')
})

it('shows no badge for humans', () => {
	expect(mountUser({username: 'human'}).find('.bot-badge').exists()).toBe(false)
})
