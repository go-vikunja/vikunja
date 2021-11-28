<template>
	<div>
		<a @click="$store.commit('menuActive', false)" class="menu-hide-button" v-if="menuActive">
			<icon icon="times" />
		</a>
		<div
			:class="{'has-background': background}"
			:style="{'background-image': background && `url(${background})`}"
			class="app-container"
		>
			<navigation/>
			<div
				:class="[
					{ 'is-menu-enabled': menuActive },
					$route.name,
				]"
				class="app-content"
			>
				<a @click="$store.commit('menuActive', false)" class="mobile-overlay" v-if="menuActive"></a>

				<quick-actions/>

				<router-view/>

				<router-view name="popup" v-slot="{ Component }">
					<transition name="modal">
						<component :is="Component" />
					</transition>
				</router-view>

				<a
					class="keyboard-shortcuts-button"
					@click="showKeyboardShortcuts()"
					v-shortcut="'?'"
				>
					<icon icon="keyboard"/>
				</a>
			</div>
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import {CURRENT_LIST, KEYBOARD_SHORTCUTS_ACTIVE, MENU_ACTIVE} from '@/store/mutation-types'
import Navigation from '@/components/home/navigation.vue'
import QuickActions from '@/components/quick-actions/quick-actions.vue'

export default {
	name: 'contentAuth',
	components: {QuickActions, Navigation},
	watch: {
		'$route': {
			handler: 'doStuffAfterRoute',
			deep: true,
		},
	},
	created() {
		this.renewTokenOnFocus()
		this.loadLabels()
	},
	computed: mapState({
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
				this.$route.name.startsWith('user.settings') ||
				this.$route.name === 'namespaces.index'
			) {
				return this.$store.dispatch(CURRENT_LIST, null)
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
		showKeyboardShortcuts() {
			this.$store.commit(KEYBOARD_SHORTCUTS_ACTIVE, true)
		},
		loadLabels() {
			this.$store.dispatch('labels/loadAllLabels')
		},
	},
}
</script>

<style lang="scss" scoped>
.menu-hide-button {
	position: fixed;
	top: 0.5rem;
	right: 0.5rem;
	z-index: 31;
	width: 3rem;
	height: 3rem;
	display: flex;
	justify-content: center;
	align-items: center;
	font-size: 2rem;
	color: var(--grey-400);
	line-height: 1;
	transition: all $transition;

	@media screen and (min-width: $tablet) {
		display: none;
	}

	&:hover,
	&:focus {
		height: 1rem;
		color: var(--grey-600);
	}
}

.app-container {
  min-height: calc(100vh - 65px);

  @media screen and (max-width: $tablet) {
    padding-top: $navbar-height;
  }

  .app-content {
    padding: $navbar-height + 1.5rem 1.5rem 1rem 1.5rem;
    z-index: 2;

    @media screen and (max-width: $tablet) {
      margin-left: 0;
      padding-top: 1.5rem;
      min-height: calc(100vh - 4rem);
    }

    &.is-menu-enabled {
      margin-left: $navbar-width;

      @media screen and (max-width: $tablet) {
        min-width: 100%;
        margin-left: 0;
      }
    }

	&.task\.detail {
		padding-left: 0;
		padding-right: 0;
	}

    .card {
      background: var(--white);
    }
  }
}

.mobile-overlay {
	display: none;
	position: fixed;
	top: 0;
	bottom: 0;
	left: 0;
	right: 0;
	background: rgba(250, 250, 250, 0.8);
	z-index: 5;
	opacity: 0;
	transition: all $transition;

	@media screen and (max-width: $tablet) {
		display: block;
		opacity: 1;
	}
}

.keyboard-shortcuts-button {
	position: fixed;
	bottom: calc(1rem - 4px);
	right: 1rem;
	z-index: 4500; // The modal has a z-index of 4000

	color: var(--grey-500);
	transition: color $transition;
	
	@media screen and (max-width: $tablet) {
		display: none;
	}
}
</style>