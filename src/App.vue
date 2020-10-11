<template>
	<div>
		<div v-if="online">
			<!-- This is a workaround to get the sw to "see" the to-be-cached version of the offline background image -->
			<div class="offline" style="height: 0;width: 0;"></div>
			<nav
				:class="{'has-background': background}"
				aria-label="main navigation"
				class="navbar main-theme is-fixed-top"
				role="navigation"
				v-if="userAuthenticated && (userInfo && userInfo.type === authTypes.USER)">
				<div class="navbar-brand">
					<router-link :to="{name: 'home'}" class="navbar-item logo">
						<img alt="Vikunja" src="/images/logo-full-pride.svg" v-if="(new Date()).getMonth() === 5"/>
						<img alt="Vikunja" src="/images/logo-full.svg" v-else/>
					</router-link>
					<a
						:class="{'is-visible': !menuActive}"
						@click="menuActive = true"
						class="menu-show-button"
					>
						<icon icon="bars"></icon>
					</a>
				</div>
				<a
					@click="menuActive = true"
					class="menu-show-button"
				>
					<icon icon="bars"></icon>
				</a>
				<div class="list-title" v-if="currentList.id">
					<h1
						:style="{ 'opacity': currentList.title === '' ? '0': '1' }"
						class="title">
						{{ currentList.title === '' ? 'Loading...' : currentList.title }}
					</h1>
					<router-link
						:to="{ name: 'list.edit', params: { id: currentList.id } }"
						class="icon"
						v-if="canWriteCurrentList">
						<icon icon="cog" size="2x"/>
					</router-link>
				</div>
				<div class="navbar-end">
					<div class="update-notification" v-if="updateAvailable">
						<p>There is an update for Vikunja available!</p>
						<a @click="refreshApp()" class="button is-primary noshadow">Update Now</a>
					</div>
					<div class="user">
						<img :src="userAvatar" alt="" class="avatar"/>
						<div class="dropdown is-right is-active">
							<div class="dropdown-trigger">
								<button @click.stop="userMenuActive = !userMenuActive" class="button noshadow">
									<span class="username">{{ userInfo.username }}</span>
									<span class="icon is-small">
									<icon icon="chevron-down"/>
								</span>
								</button>
							</div>
							<transition name="fade">
								<div class="dropdown-menu" v-if="userMenuActive">
									<div class="dropdown-content">
										<router-link :to="{name: 'user.settings'}" class="dropdown-item">
											Settings
										</router-link>
										<a :href="imprintUrl" class="dropdown-item" target="_blank" v-if="imprintUrl">Imprint</a>
										<a
											:href="privacyPolicyUrl"
											class="dropdown-item"
											target="_blank"
											v-if="privacyPolicyUrl">
											Privacy policy
										</a>
										<a @click="keyboardShortcutsActive = true" class="dropdown-item">Keyboard
											Shortcuts</a>
										<a @click="logout()" class="dropdown-item">
											Logout
										</a>
									</div>
								</div>
							</transition>
						</div>
					</div>
				</div>
			</nav>
			<div v-if="userAuthenticated && (userInfo && userInfo.type === authTypes.USER)">
				<a @click="menuActive = false" class="menu-hide-button" v-if="menuActive">
					<icon icon="times"></icon>
				</a>
				<div
					:class="{'has-background': background}"
					:style="{'background-image': `url(${background})`}"
					class="app-container"
				>
					<div :class="{'is-active': menuActive}" class="namespace-container">
						<div class="menu top-menu">
							<router-link :to="{name: 'home'}" class="logo">
								<img alt="Vikunja" src="/images/logo-full.svg"/>
							</router-link>
							<ul class="menu-list">
								<li>
									<router-link :to="{ name: 'home'}">
									<span class="icon">
										<icon icon="calendar"/>
									</span>
										Overview
									</router-link>
								</li>
								<li>
									<router-link :to="{ name: 'tasks.range', params: {type: 'week'}}">
									<span class="icon">
										<icon icon="calendar-week"/>
									</span>
										Next Week
									</router-link>
								</li>
								<li>
									<router-link :to="{ name: 'tasks.range', params: {type: 'month'}}">
									<span class="icon">
										<icon :icon="['far', 'calendar-alt']"/>
									</span>
										Next Month
									</router-link>
								</li>
								<li>
									<router-link :to="{ name: 'teams.index'}">
									<span class="icon">
										<icon icon="users"/>
									</span>
										Teams
									</router-link>
								</li>
								<li>
									<router-link :to="{ name: 'namespaces.index'}">
									<span class="icon">
										<icon icon="layer-group"/>
									</span>
										Namespaces & Lists
									</router-link>
								</li>
								<li>
									<router-link :to="{ name: 'labels.index'}">
									<span class="icon">
										<icon icon="tags"/>
									</span>
										Labels
									</router-link>
								</li>
							</ul>
						</div>

						<a
							@click="menuActive = false"
							@shortkey="() => menuActive = !menuActive"
							class="collapse-menu-button"
							v-shortkey="['ctrl', 'e']">
							Collapse Menu
						</a>

						<aside class="menu namespaces-lists">
							<template v-for="n in namespaces">
								<div :key="n.id">
									<router-link
										:to="{name: 'namespace.edit', params: {id: n.id} }"
										class="nsettings"
										v-if="n.id > 0"
										v-tooltip.right="'Settings'">
										<span class="icon">
											<icon icon="cog"/>
										</span>
									</router-link>
									<router-link
										:key="n.id + 'list.create'"
										:to="{ name: 'list.create', params: { id: n.id} }"
										class="nsettings"
										v-if="n.id > 0"
										v-tooltip="'Add a new list in the ' + n.title + ' namespace'">
										<span class="icon">
											<icon icon="plus"/>
										</span>
									</router-link>
									<label
										:for="n.id + 'checker'"
										class="menu-label"
										v-tooltip="n.title + ' (' + n.lists.length + ')'">
										<span class="name">
											<span
												:style="{ backgroundColor: n.hexColor }"
												class="color-bubble"
												v-if="n.hexColor !== ''">
											</span>
											{{ n.title }} ({{ n.lists.length }})
										</span>
									</label>
								</div>
								<input
									:id="n.id + 'checker'"
									:key="n.id + 'checker'"
									checked="checked"
									class="checkinput"
									type="checkbox"/>
								<div :key="n.id + 'child'" class="more-container">
									<ul class="menu-list can-be-hidden">
										<template v-for="l in n.lists">
											<!-- This is a bit ugly but vue wouldn't want to let me filter this - probably because the lists
													are nested inside of the namespaces makes it a lot harder.-->
											<li :key="l.id" v-if="!l.isArchived">
												<router-link
													class="list-menu-link"
													:class="{'router-link-exact-active': currentList.id === l.id}"
													:to="{ name: 'list.index', params: { listId: l.id} }"
													tag="span"
												>
													<span
														:style="{ backgroundColor: l.hexColor }"
														class="color-bubble"
														v-if="l.hexColor !== ''">
													</span>
													<span class="list-menu-title">
														{{ l.title }}
													</span>
													<span
														:class="{'is-favorite': l.isFavorite}"
														@click.stop="toggleFavoriteList(l)"
														class="favorite">
														<icon icon="star" v-if="l.isFavorite"/>
														<icon :icon="['far', 'star']" v-else/>
													</span>
												</router-link>
											</li>
										</template>
									</ul>
									<label :for="n.id + 'checker'" class="hidden-hint">
										Show hidden lists ({{ n.lists.length }})...
									</label>
								</div>
							</template>
						</aside>
						<a class="menu-bottom-link" href="https://vikunja.io" target="_blank">Powered by Vikunja</a>
					</div>
					<div
						:class="[
								{
									'fullpage-overlay': fullpage,
									'is-menu-enabled': menuActive,
								},
								$route.name,
							]"
						class="app-content"
					>
						<a @click="menuActive = false" class="mobile-overlay" v-if="menuActive"></a>
						<transition name="fade">
							<router-view/>
						</transition>
						<a @click="keyboardShortcutsActive = true" class="keyboard-shortcuts-button">
							<icon icon="keyboard"/>
						</a>
					</div>
				</div>
			</div>
			<div
				:class="{'has-background': background}"
				:style="{'background-image': `url(${background})`}"
				class="link-share-container"
				v-else-if="userAuthenticated && (userInfo && userInfo.type === authTypes.LINK_SHARE)"
			>
				<div class="container has-text-centered link-share-view">
					<div class="column is-10 is-offset-1">
						<img alt="Vikunja" class="logo" src="/images/logo-full.svg"/>
						<h1
							:style="{ 'opacity': currentList.title === '' ? '0': '1' }"
							class="title">
							{{ currentList.title === '' ? 'Loading...' : currentList.title }}
						</h1>
						<div class="box has-text-left view">
							<div class="logout">
								<a @click="logout()" class="button">
									<span>Logout</span>
									<span class="icon is-small">
									<icon icon="sign-out-alt"/>
								</span>
								</a>
							</div>
							<router-view/>
						</div>
					</div>
				</div>
			</div>
			<div v-else>
				<div class="noauth-container">
					<img alt="Vikunja" src="/images/logo-full.svg"/>
					<div class="message is-info" v-if="motd !== ''">
						<div class="message-header">
							<p>Info</p>
						</div>
						<div class="message-body">
							{{ motd }}
						</div>
					</div>
					<router-view/>
				</div>
			</div>
			<notification/>
		</div>
		<div class="app offline" v-else>
			<div class="offline-message">
				<h1>You are offline.</h1>
				<p>Please check your network connection and try again.</p>
			</div>
		</div>

		<transition name="fade">
			<keyboard-shortcuts @close="keyboardShortcutsActive = false" v-if="keyboardShortcutsActive"/>
		</transition>
	</div>
</template>

<script>
import router from './router'
import {mapState} from 'vuex'

import authTypes from './models/authTypes'
import Rights from './models/rights.json'

import swEvents from './ServiceWorker/events'
import Notification from './components/misc/notification'
import {CURRENT_LIST, IS_FULLPAGE, ONLINE} from './store/mutation-types'
import KeyboardShortcuts from './components/misc/keyboard-shortcuts'

export default {
	name: 'app',
	components: {
		KeyboardShortcuts,
		Notification,
	},
	data() {
		return {
			menuActive: true,
			currentDate: new Date(),
			userMenuActive: false,
			authTypes: authTypes,
			keyboardShortcutsActive: false,

			// Service Worker stuff
			updateAvailable: false,
			registration: null,
			refreshing: false,
		}
	},
	beforeMount() {
		// Check if the user is offline, show a message then
		this.$store.commit(ONLINE, navigator.onLine)
		window.addEventListener('online', () => this.$store.commit(ONLINE, navigator.onLine))
		window.addEventListener('offline', () => this.$store.commit(ONLINE, navigator.onLine))

		// Password reset
		if (this.$route.query.userPasswordReset !== undefined) {
			localStorage.removeItem('passwordResetToken') // Delete an eventually preexisting old token
			localStorage.setItem('passwordResetToken', this.$route.query.userPasswordReset)
			router.push({name: 'user.password-reset.reset'})
		}
		// Email verification
		if (this.$route.query.userEmailConfirm !== undefined) {
			localStorage.removeItem('emailConfirmToken') // Delete an eventually preexisting old token
			localStorage.setItem('emailConfirmToken', this.$route.query.userEmailConfirm)
			router.push({name: 'user.login'})
		}
	},
	beforeCreate() {
		this.$store.dispatch('config/update')
		this.$store.dispatch('auth/checkAuth')
			.then(() => {
				// Check if the user is already logged in, if so, redirect them to the homepage
				if (
					!this.userAuthenticated &&
					this.$route.name !== 'user.login' &&
					this.$route.name !== 'user.password-reset.request' &&
					this.$route.name !== 'user.password-reset.reset' &&
					this.$route.name !== 'user.register' &&
					this.$route.name !== 'link-share.auth'
				) {
					router.push({name: 'user.login'})
				}

				if (this.userAuthenticated && this.userInfo.type === authTypes.USER && (this.$route.params.name === 'home' || this.namespaces.length === 0)) {
					this.loadNamespaces()
				}
			})
	},
	created() {

		// Make sure to always load the home route when running with electron
		if(this.$route.fullPath.endsWith('frontend/index.html')) {
			this.$router.push({name: 'home'})
		}

		// Service worker communication
		document.addEventListener(swEvents.SW_UPDATED, this.showRefreshUI, {once: true})

		if (navigator && navigator.serviceWorker) {
			navigator.serviceWorker.addEventListener(
				'controllerchange', () => {
					if (this.refreshing) return
					this.refreshing = true
					window.location.reload()
				},
			)
		}

		// Hide the menu by default on mobile
		if (window.innerWidth < 770) {
			this.menuActive = false
		}

		// Try renewing the token every time vikunja is loaded initially
		// (When opening the browser the focus event is not fired)
		this.$store.dispatch('auth/renewToken')

		// Check if the token is still valid if the window gets focus again to maybe renew it
		window.addEventListener('focus', () => {

			if (!this.userAuthenticated) {
				return
			}

			const expiresIn = this.userInfo.exp - +new Date() / 1000

			// If the token expiry is negative, it is already expired and we have no choice but to redirect
			// the user to the login page
			if (expiresIn < 0) {
				this.$store.dispatch('auth/checkAuth')
				router.push({name: 'user.login'})
				return
			}

			// Check if the token is valid for less than 60 hours and renew if thats the case
			if (expiresIn < 60 * 3600) {
				this.$store.dispatch('auth/renewToken')
				console.log('renewed token')
			}
		})

		// This will hide the menu once clicked outside of it
		this.$nextTick(() => document.addEventListener('click', () => this.userMenuActive = false))
	},
	watch: {
		// call the method again if the route changes
		'$route': 'doStuffAfterRoute',
	},
	computed: mapState({
		userInfo: state => state.auth.info,
		userAvatar: state => state.auth.avatarUrl,
		userAuthenticated: state => state.auth.authenticated,
		motd: state => state.config.motd,
		online: ONLINE,
		fullpage: IS_FULLPAGE,
		namespaces(state) {
			return state.namespaces.namespaces.filter(n => !n.isArchived)
		},
		currentList: CURRENT_LIST,
		background: 'background',
		imprintUrl: state => state.config.legal.imprintUrl,
		privacyPolicyUrl: state => state.config.legal.privacyPolicyUrl,
		canWriteCurrentList: state => state.currentList.maxRight > Rights.READ,
	}),
	methods: {
		logout() {
			this.$store.dispatch('auth/logout')
			router.push({name: 'user.login'})
		},
		loadNamespaces() {
			this.$store.dispatch('namespaces/loadNamespaces')
		},
		loadNamespacesIfNeeded(e) {
			if (this.userAuthenticated && (this.userInfo && this.userInfo.type === authTypes.USER) && (e.name === 'home' || this.namespaces.length === 0)) {
				this.loadNamespaces()
			}
		},
		doStuffAfterRoute(e) {
			// this.setTitle('') // Reset the title if the page component does not set one itself

			if (this.$store.state[IS_FULLPAGE]) {
				this.$store.commit(IS_FULLPAGE, false)
			}

			this.loadNamespacesIfNeeded(e)
			this.userMenuActive = false

			// If the menu is active on desktop, don't hide it because that would confuse the user
			if (window.innerWidth < 770) {
				this.menuActive = false
			}

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
		showRefreshUI(e) {
			console.log('recieved refresh event', e)
			this.registration = e.detail
			this.updateAvailable = true
		},
		refreshApp() {
			this.updateExists = false
			if (!this.registration || !this.registration.waiting) {
				return
			}
			// Notify the service worker to actually do the update
			this.registration.waiting.postMessage('skipWaiting')
		},
		toggleFavoriteList(list) {
			// The favorites pseudo list is always favorite
			// Archived lists cannot be marked favorite
			if (list.id === -1 || list.isArchived) {
				return
			}
			this.$store.dispatch('lists/toggleListFavorite', list)
				.catch(e => this.error(e, this))
		},
	},
}
</script>
