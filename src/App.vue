<template>
	<div>
		<div v-if="online">
			<!-- This is a workaround to get the sw to "see" the to-be-cached version of the offline background image -->
			<div class="offline" style="height: 0;width: 0;"></div>
			<nav
					class="navbar main-theme is-fixed-top"
					:class="{'has-background': background}"
					role="navigation"
					aria-label="main navigation"
					v-if="userAuthenticated && (userInfo && userInfo.type === authTypes.USER)">
				<div class="navbar-brand">
					<router-link :to="{name: 'home'}" class="navbar-item logo">
						<img src="/images/logo-full-pride.svg" alt="Vikunja" v-if="(new Date()).getMonth() === 5"/>
						<img src="/images/logo-full.svg" alt="Vikunja" v-else/>
					</router-link>
					<a
							@click="menuActive = true"
							class="menu-show-button"
							:class="{'is-visible': !menuActive}"
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
							class="title"
							:style="{ 'opacity': currentList.title === '' ? '0': '1' }">
						{{ currentList.title === '' ? 'Loading...': currentList.title}}
					</h1>
					<router-link :to="{ name: 'list.edit', params: { id: currentList.id } }" class="icon">
						<icon icon="cog" size="2x"/>
					</router-link>
				</div>
				<div class="navbar-end">
					<div v-if="updateAvailable" class="update-notification">
						<p>There is an update for Vikunja available!</p>
						<a @click="refreshApp()" class="button is-primary noshadow">Update Now</a>
					</div>
					<div class="user">
						<img :src="userInfo.getAvatarUrl()" class="avatar" alt=""/>
						<div class="dropdown is-right is-active">
							<div class="dropdown-trigger">
								<button class="button noshadow" @click="userMenuActive = !userMenuActive">
									<span class="username">{{userInfo.username}}</span>
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
						class="app-container"
						:class="{'has-background': background}"
						:style="{'background-image': `url(${background})`}"
				>
					<div class="namespace-container" :class="{'is-active': menuActive}">
						<div class="menu top-menu">
							<router-link :to="{name: 'home'}" class="logo">
								<img src="/images/logo-full.svg" alt="Vikunja"/>
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
						<a @click="menuActive = false" class="collapse-menu-button">Collapse Menu</a>
						<aside class="menu namespaces-lists">
							<div class="spinner" :class="{ 'is-loading': namespaceService.loading}"></div>
							<template v-for="n in namespaces">
								<div :key="n.id">
									<router-link
											v-tooltip.right="'Settings'"
											:to="{name: 'namespace.edit', params: {id: n.id} }"
											class="nsettings"
											v-if="n.id > 0">
										<span class="icon">
											<icon icon="cog"/>
										</span>
									</router-link>
									<router-link
											v-tooltip="'Add a new list in the ' + n.title + ' namespace'"
											:to="{ name: 'list.create', params: { id: n.id} }"
											class="nsettings"
											:key="n.id + 'list.create'"
											v-if="n.id > 0">
										<span class="icon">
											<icon icon="plus"/>
										</span>
									</router-link>
									<label
											class="menu-label"
											v-tooltip="n.title + ' (' + n.lists.length + ')'"
											:for="n.id + 'checker'">
										<span class="name">
											<span
													class="color-bubble"
													v-if="n.hexColor !== ''"
													:style="{ backgroundColor: n.hexColor }">
											</span>
											{{n.title}} ({{n.lists.length}})
										</span>
									</label>
								</div>
								<input
										:key="n.id + 'checker'"
										type="checkbox"
										checked="checked"
										:id="n.id + 'checker'"
										class="checkinput"/>
								<div class="more-container" :key="n.id + 'child'">
									<ul class="menu-list can-be-hidden">
										<li v-for="l in n.lists" :key="l.id">
											<router-link
													:to="{ name: 'list.index', params: { listId: l.id} }"
													:class="{'router-link-exact-active': currentList.id === l.id}">
												<span class="name">
													<span
															class="color-bubble"
															v-if="l.hexColor !== ''"
															:style="{ backgroundColor: l.hexColor }">
													</span>
													{{l.title}}
												</span>
											</router-link>
										</li>
									</ul>
									<label class="hidden-hint" :for="n.id + 'checker'">
										Show hidden lists ({{n.lists.length}})...
									</label>
								</div>
							</template>
						</aside>
						<a class="menu-bottom-link" target="_blank" href="https://vikunja.io">Powered by Vikunja</a>
					</div>
					<div
							class="app-content"
							:class="{
								'fullpage-overlay': fullpage,
								'is-menu-enabled': menuActive,
							}"
					>
						<a class="mobile-overlay" v-if="menuActive" @click="menuActive = false"></a>
						<transition name="fade">
							<router-view/>
						</transition>
					</div>
				</div>
			</div>
			<div
					v-else-if="userAuthenticated && (userInfo && userInfo.type === authTypes.LINK_SHARE)"
					class="link-share-container"
					:class="{'has-background': background}"
					:style="{'background-image': `url(${background})`}"
			>
				<div class="container has-text-centered link-share-view">
					<div class="column is-10 is-offset-1">
						<img src="/images/logo-full.svg" alt="Vikunja" class="logo"/>
						<h1
								class="title"
								:style="{ 'opacity': currentList.title === '' ? '0': '1' }">
							{{ currentList.title === '' ? 'Loading...': currentList.title}}
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
					<img src="/images/logo-full.svg" alt="Vikunja"/>
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
	</div>
</template>

<script>
	import router from './router'
	import {mapState} from 'vuex'

	import NamespaceService from './services/namespace'
	import authTypes from './models/authTypes'

	import swEvents from './ServiceWorker/events'
	import Notification from './components/misc/notification'
	import {CURRENT_LIST, IS_FULLPAGE, ONLINE} from './store/mutation-types'

	export default {
		name: 'app',
		components: {
			Notification,
		},
		data() {
			return {
				namespaceService: NamespaceService,
				menuActive: true,
				currentDate: new Date(),
				userMenuActive: false,
				authTypes: authTypes,

				// Service Worker stuff
				updateAvailable: false,
				registration: null,
				refreshing: false,
			}
		},
		beforeMount() {
			// Check if the user is offline, show a message then
			this.$store.commit(ONLINE, navigator.onLine)
			window.addEventListener('online', () => this.$store.commit(ONLINE, navigator.onLine));
			window.addEventListener('offline', () => this.$store.commit(ONLINE, navigator.onLine));

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

			// Service worker communication
			document.addEventListener(swEvents.SW_UPDATED, this.showRefreshUI, {once: true})

			if (navigator && navigator.serviceWorker) {
				navigator.serviceWorker.addEventListener(
					'controllerchange', () => {
						if (this.refreshing) return;
						this.refreshing = true;
						window.location.reload();
					}
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
		},
		watch: {
			// call the method again if the route changes
			'$route': 'doStuffAfterRoute',
		},
		computed: mapState({
			userInfo: state => state.auth.info,
			userAuthenticated: state => state.auth.authenticated,
			motd: state => state.config.motd,
			online: ONLINE,
			fullpage: IS_FULLPAGE,
			namespaces(state) {
				return state.namespaces.namespaces.filter(n => !n.isArchived)
			},
			currentList: CURRENT_LIST,
			background: 'background',
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
				this.registration = e.detail;
				this.updateAvailable = true;
			},
			refreshApp() {
				this.updateExists = false;
				if (!this.registration || !this.registration.waiting) {
					return;
				}
				// Notify the service worker to actually do the update
				this.registration.waiting.postMessage('skipWaiting');
			},
		},
	}
</script>
