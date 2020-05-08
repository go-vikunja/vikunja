<template>
	<div>
		<div v-if="online">
			<!-- This is a workaround to get the sw to "see" the to-be-cached version of the offline background image -->
			<div class="offline" style="height: 0;width: 0;"></div>
			<nav class="navbar main-theme is-fixed-top" role="navigation" aria-label="main navigation"
				v-if="userAuthenticated && (userInfo && userInfo.type === authTypes.USER)">
				<div class="navbar-brand">
					<router-link :to="{name: 'home'}" class="navbar-item logo">
						<img src="/images/logo-full.svg" alt="Vikunja"/>
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
										<router-link :to="{name: 'userSettings'}" class="dropdown-item">
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
				<a @click="mobileMenuActive = true" class="mobilemenu-show-button" v-if="!mobileMenuActive">
					<icon icon="bars"></icon>
				</a>
				<a @click="mobileMenuActive = false" class="mobilemenu-hide-button" v-if="mobileMenuActive">
					<icon icon="times"></icon>
				</a>
				<div class="app-container">
					<div class="namespace-container" :class="{'is-active': mobileMenuActive}">
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
									<router-link :to="{ name: 'showTasksInRange', params: {type: 'week'}}">
									<span class="icon">
										<icon icon="calendar-week"/>
									</span>
										Next Week
									</router-link>
								</li>
								<li>
									<router-link :to="{ name: 'showTasksInRange', params: {type: 'month'}}">
									<span class="icon">
										<icon :icon="['far', 'calendar-alt']"/>
									</span>
										Next Month
									</router-link>
								</li>
								<li>
									<router-link :to="{ name: 'listTeams'}">
									<span class="icon">
										<icon icon="users"/>
									</span>
										Teams
									</router-link>
								</li>
								<li>
									<router-link :to="{ name: 'newNamespace'}">
									<span class="icon">
										<icon icon="layer-group"/>
									</span>
										New Namespace
									</router-link>
								</li>
								<li>
									<router-link :to="{ name: 'listLabels'}">
									<span class="icon">
										<icon icon="tags"/>
									</span>
										Labels
									</router-link>
								</li>
							</ul>
						</div>
						<aside class="menu namespaces-lists">
							<fancycheckbox v-model="showArchived" class="show-archived-check">
								Show Archived
							</fancycheckbox>
							<div class="spinner" :class="{ 'is-loading': namespaceService.loading}"></div>
							<template v-for="n in namespaces">
								<div :key="n.id">
									<router-link v-tooltip.right="'Settings'"
											:to="{name: 'editNamespace', params: {id: n.id} }" class="nsettings"
											v-if="n.id > 0">
										<span class="icon">
											<icon icon="cog"/>
										</span>
									</router-link>
									<router-link v-tooltip="'Add a new list in the ' + n.name + ' namespace'"
											:to="{ name: 'newList', params: { id: n.id} }" class="nsettings"
											:key="n.id + 'newList'" v-if="n.id > 0">
										<span class="icon">
											<icon icon="plus"/>
										</span>
									</router-link>
									<label class="menu-label" v-tooltip="n.name + ' (' + n.lists.length + ')'" :for="n.id + 'checker'">
										<span class="name">
											<span class="color-bubble" v-if="n.hexColor !== ''" :style="{ backgroundColor: n.hexColor }"></span>
											{{n.name}} ({{n.lists.length}})
										</span>
										<span class="is-archived" v-if="n.isArchived">
											Archived
										</span>
									</label>
								</div>
								<input :key="n.id + 'checker'" type="checkbox" checked="checked" :id="n.id + 'checker'" class="checkinput"/>
								<div class="more-container" :key="n.id + 'child'">
									<ul class="menu-list can-be-hidden" >
										<li v-for="l in n.lists" :key="l.id">
											<router-link :to="{ name: 'list.index', params: { listId: l.id} }">
												<span class="name">
													<span class="color-bubble" v-if="l.hexColor !== ''" :style="{ backgroundColor: l.hexColor }"></span>
													{{l.title}}
												</span>
												<span class="is-archived" v-if="l.isArchived">
													Archived
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
					<div class="app-content" :class="{'fullpage-overlay': fullpage}">
						<a class="mobile-overlay" v-if="mobileMenuActive" @click="mobileMenuActive = false"></a>
						<transition name="fade">
							<router-view/>
						</transition>
					</div>
				</div>
			</div>
			<div v-else-if="userAuthenticated && (userInfo && userInfo.type === authTypes.LINK_SHARE)">
				<div class="container has-text-centered link-share-view">
					<div class="column is-10 is-offset-1">
						<img src="/images/logo-full.svg" alt="Vikunja" class="logo"/>
						<div class="box has-text-left">
							<div class="logout">
								<a @click="logout()" class="button logout">
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
				<div class="container">
					<div class="column is-4 is-offset-4">
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
	import Notification from './components/global/notification'
	import Fancycheckbox from './components/global/fancycheckbox'
	import {IS_FULLPAGE, ONLINE} from './store/mutation-types'

	export default {
		name: 'app',
		components: {
			Fancycheckbox,
			Notification,
		},
		data() {
			return {
				namespaceService: NamespaceService,
				mobileMenuActive: false,
				currentDate: new Date(),
				userMenuActive: false,
				authTypes: authTypes,
				showArchived: false,

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
				router.push({name: 'passwordReset'})
			}
			// Email verification
			if (this.$route.query.userEmailConfirm !== undefined) {
				localStorage.removeItem('emailConfirmToken') // Delete an eventually preexisting old token
				localStorage.setItem('emailConfirmToken', this.$route.query.userEmailConfirm)
				router.push({name: 'login'})
			}
		},
		created() {
			this.$store.dispatch('config/update')
			this.$store.dispatch('auth/checkAuth')

			if (this.userAuthenticated && this.userInfo.type === authTypes.USER && (this.$route.params.name === 'home' || this.namespaces.length === 0)) {
				this.loadNamespaces()
			}

			// Service worker communication
			document.addEventListener(swEvents.SW_UPDATED, this.showRefreshUI, {once: true})

			navigator.serviceWorker.addEventListener(
				'controllerchange', () => {
					if (this.refreshing) return;
					this.refreshing = true;
					window.location.reload();
				}
			);

			// Schedule a token renew every minute
			setTimeout(() => {
				this.$store.dispatch('auth/renewToken')
			}, 1000 * 60)
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
				return state.namespaces.namespaces.filter(n => this.showArchived ? true : !n.isArchived)
			},
		}),
		methods: {
			logout() {
				this.$store.dispatch('auth/logout')
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
				this.$store.commit(IS_FULLPAGE, false)
				this.loadNamespacesIfNeeded(e)
				this.mobileMenuActive = false
				this.userMenuActive = false
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
