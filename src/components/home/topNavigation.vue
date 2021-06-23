<template>
	<nav
		:class="{'has-background': background}"
		aria-label="main navigation"
		class="navbar main-theme is-fixed-top"
		role="navigation"
	>
		<div class="navbar-brand">
			<router-link :to="{name: 'home'}" class="navbar-item logo">
				<img alt="Vikunja" src="/images/logo-full-pride.svg" v-if="(new Date()).getMonth() === 5"/>
				<img alt="Vikunja" src="/images/logo-full.svg" v-else/>
			</router-link>
			<a
				@click="$store.commit('toggleMenu')"
				class="menu-show-button"
				@shortkey="() => $store.commit('toggleMenu')"
				v-shortkey="['ctrl', 'e']"
			>
				<icon icon="bars"></icon>
			</a>
		</div>
		<a
			@click="$store.commit('toggleMenu')"
			class="menu-show-button"
		>
			<icon icon="bars"></icon>
		</a>
		<div class="list-title" v-if="currentList.id" ref="listTitle">
			<h1
				:style="{ 'opacity': currentList.title === '' ? '0': '1' }"
				class="title">
				{{ currentList.title === '' ? 'Loading...' : currentList.title }}
			</h1>

			<list-settings-dropdown v-if="canWriteCurrentList && currentList.id !== -1" :list="currentList"/>
		</div>

		<div class="navbar-end">
			<update/>
			<a
				@click="openQuickActions"
				class="trigger-button pr-0"
				@shortkey="openQuickActions"
				v-shortkey="['ctrl', 'k']"
			>
				<icon icon="search"/>
			</a>
			<notifications/>
			<div class="user">
				<img :src="userAvatar" alt="" class="avatar"/>
				<dropdown class="is-right" ref="usernameDropdown">
					<template v-slot:trigger>
						<x-button
							type="secondary"
							:shadow="false">
							<span class="username">{{ userInfo.name !== '' ? userInfo.name : userInfo.username }}</span>
							<span class="icon is-small">
								<icon icon="chevron-down"/>
							</span>
						</x-button>
					</template>

					<router-link :to="{name: 'user.settings'}" class="dropdown-item">
						Settings
					</router-link>
					<a
						:href="imprintUrl"
						class="dropdown-item"
						target="_blank"
						v-if="imprintUrl">
						Imprint
					</a>
					<a
						:href="privacyPolicyUrl"
						class="dropdown-item"
						target="_blank"
						v-if="privacyPolicyUrl">
						Privacy policy
					</a>
					<a @click="$store.commit('keyboardShortcutsActive', true)" class="dropdown-item">
						Keyboard Shortcuts
					</a>
					<a @click="logout()" class="dropdown-item">
						Logout
					</a>
				</dropdown>
			</div>
		</div>
	</nav>
</template>

<script>
import {mapState} from 'vuex'
import {CURRENT_LIST, QUICK_ACTIONS_ACTIVE} from '@/store/mutation-types'
import Rights from '@/models/rights.json'
import Update from '@/components/home/update'
import ListSettingsDropdown from '@/components/list/list-settings-dropdown'
import Dropdown from '@/components/misc/dropdown'
import Notifications from '@/components/notifications/notifications'

export default {
	name: 'topNavigation',
	components: {
		Notifications,
		Dropdown,
		ListSettingsDropdown,
		Update,
	},
	computed: mapState({
		userInfo: state => state.auth.info,
		userAvatar: state => state.auth.avatarUrl,
		userAuthenticated: state => state.auth.authenticated,
		currentList: CURRENT_LIST,
		background: 'background',
		imprintUrl: state => state.config.legal.imprintUrl,
		privacyPolicyUrl: state => state.config.legal.privacyPolicyUrl,
		canWriteCurrentList: state => state.currentList.maxRight > Rights.READ,
	}),
	mounted() {
		const usernameWidth = this.$refs.usernameDropdown.$el.clientWidth
		this.$refs.listTitle.style.setProperty('--nav-username-width', `${usernameWidth}px`)
	},
	methods: {
		logout() {
			this.$store.dispatch('auth/logout')
			this.$router.push({name: 'user.login'})
		},
		openQuickActions() {
			this.$store.commit(QUICK_ACTIONS_ACTIVE, true)
		},
	},
}
</script>
