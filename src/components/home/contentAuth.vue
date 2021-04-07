<template>
	<div>
		<a @click="$store.commit('menuActive', false)" class="menu-hide-button" v-if="menuActive">
			<icon icon="times"></icon>
		</a>
		<div
			:class="{'has-background': background}"
			:style="{'background-image': `url(${background})`}"
			class="app-container"
		>
			<navigation/>
			<div
				:class="[
					{
						'is-menu-enabled': menuActive,
					},
					$route.name,
				]"
				class="app-content"
			>
				<a @click="$store.commit('menuActive', false)" class="mobile-overlay" v-if="menuActive"></a>
				
				<router-view/>
				
				<transition name="modal">
					<router-view name="popup"/>
				</transition>
				
				<a @click="$store.commit('keyboardShortcutsActive', true)" class="keyboard-shortcuts-button">
					<icon icon="keyboard"/>
				</a>
			</div>
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import {CURRENT_LIST, MENU_ACTIVE} from '@/store/mutation-types'
import Navigation from '@/components/home/navigation'

export default {
	name: 'contentAuth',
	components: {Navigation},
	watch: {
		'$route': 'doStuffAfterRoute',
	},
	created() {
		this.renewTokenOnFocus()
	},
	computed: mapState({
		namespaces(state) {
			return state.namespaces.namespaces.filter(n => !n.isArchived)
		},
		currentList: CURRENT_LIST,
		background: 'background',
		menuActive: MENU_ACTIVE,
		userInfo: state => state.auth.info,
		authenticated: state => state.auth.authenticated,
	}),
	methods: {
		doStuffAfterRoute() {
			// this.setTitle('') // Reset the title if the page component does not set one itself
			this.hideMenuOnMobile()
			this.resetCurrentList()
		},
		resetCurrentList() {
			// Reset the current list highlight in menu if the current list is not list related.
			if (
				this.$route.name === 'home' ||
				this.$route.name === 'namespace.edit' ||
				this.$route.name === 'teams.index' ||
				this.$route.name === 'teams.edit' ||
				this.$route.name === 'tasks.range' ||
				this.$route.name === 'labels.index' ||
				this.$route.name === 'migrate.start' ||
				this.$route.name === 'migrate.wunderlist' ||
				this.$route.name === 'user.settings' ||
				this.$route.name === 'namespaces.index'
			) {
				this.$store.commit(CURRENT_LIST, {})
			}
		},
		renewTokenOnFocus() {
			// Try renewing the token every time vikunja is loaded initially
			// (When opening the browser the focus event is not fired)
			this.$store.dispatch('auth/renewToken')

			// Check if the token is still valid if the window gets focus again to maybe renew it
			window.addEventListener('focus', () => {

				if (!this.authenticated) {
					return
				}

				const expiresIn = (this.userInfo !== null ? this.userInfo.exp : 0) - +new Date() / 1000

				// If the token expiry is negative, it is already expired and we have no choice but to redirect
				// the user to the login page
				if (expiresIn < 0) {
					this.$store.dispatch('auth/checkAuth')
					this.$router.push({name: 'user.login'})
					return
				}

				// Check if the token is valid for less than 60 hours and renew if thats the case
				if (expiresIn < 60 * 3600) {
					this.$store.dispatch('auth/renewToken')
					console.debug('renewed token')
				}
			})
		},
		hideMenuOnMobile() {
			if (window.innerWidth < 769) {
				this.$store.commit(MENU_ACTIVE, false)
			}
		},
	},
}
</script>
