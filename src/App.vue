<template>
	<div id="app">
		<nav class="navbar is-dark is-fixed-top" role="navigation" aria-label="main navigation" v-if="user.authenticated">
			<div class="navbar-brand">
				<router-link :to="{name: 'home'}" class="navbar-item logo">
					<img src="/images/logo-full-white.svg"/>
				</router-link>
			</div>
		</nav>
		<div v-if="user.authenticated">
			<a @click="mobileMenuActive = true" class="mobilemenu-show-button" v-if="!mobileMenuActive"><icon icon="bars"></icon></a>
			<a @click="mobileMenuActive = false" class="mobilemenu-hide-button" v-if="mobileMenuActive"><icon icon="times"></icon></a>
			<div class="app-container">
				<div class="namespace-container" :class="{'is-active': mobileMenuActive}">
					<div class="menu top-menu">
						<ul class="menu-list user">
							<li>
								<img :src="gravatar()" class="is-rounded" alt=""/>
								<span class="username">{{user.infos.username}}</span>
								<a @click="logout()" class="logout-icon">
									<span class="icon is-medium">
										<icon icon="power-off" size="2x"/>
									</span>
								</a>
							</li>
						</ul>
					</div>
					<div class="menu top-menu">
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
								<router-link :to="{ name: 'showTasksInRange', params: {type: 'month'}}">
									<span class="icon">
										<icon :icon="['far', 'calendar-alt']"/>
									</span>
									Next Month
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
						</ul>
					</div>
					<aside class="menu namespaces-lists">
						<div class="spinner" :class="{ 'is-loading': loading}"></div>
						<template v-for="n in namespaces">
							<div :key="n.id">
								<router-link v-tooltip.right="'Settings'" :to="{name: 'editNamespace', params: {id: n.id} }" class="nsettings" v-if="n.id > 0">
										<span class="icon">
											<icon icon="cog"/>
										</span>
								</router-link>
								<router-link v-tooltip="'Add a new list in the ' + n.name + ' namespace'" :to="{ name: 'newList', params: { id: n.id} }" class="nsettings" :key="n.id + 'newList'" v-if="n.id > 0">
										<span class="icon">
											<icon icon="plus"/>
										</span>
								</router-link>
								<div class="menu-label">
									{{n.name}}
								</div>
							</div>
							<ul class="menu-list" :key="n.id + 'child'">
								<li v-for="l in n.lists" :key="l.id">
									<router-link :to="{ name: 'showList', params: { id: l.id} }">{{l.title}}</router-link>
								</li>
							</ul>
						</template>
					</aside>
				</div>
				<div class="app-content" :class="{'fullpage-overlay': fullpage}">
					<a class="mobile-overlay" v-if="mobileMenuActive" @click="mobileMenuActive = false"></a>
					<transition name="fade">
						<router-view/>
					</transition>
				</div>
			</div>
		</div>
		<div v-else>
			<div class="container has-text-centered">
				<div class="column is-4 is-offset-4">
					<img src="/images/logo-full.svg"/>
						<router-view/>
				</div>
			</div>
		</div>
	<notifications position="bottom left" />
	</div>
</template>

<script>
    import auth from './auth'
    import {HTTP} from './http-common'
	import message from './message'
    import router from './router'

    export default {
        name: 'app',

        data() {
            return {
                user: auth.user,
                loading: false,
                namespaces: [],
				mobileMenuActive: false,
				fullpage: false,
				currentDate: new Date(),
            }
        },
		beforeMount() {
            // Password reset
            if(this.$route.query.userPasswordReset !== undefined) {
				localStorage.removeItem('passwordResetToken') // Delete an eventually preexisting old token
				localStorage.setItem('passwordResetToken', this.$route.query.userPasswordReset)
                router.push({name: 'passwordReset'})
			}
            // Email verification
            if(this.$route.query.userEmailConfirm !== undefined) {
                localStorage.removeItem('emailConfirmToken') // Delete an eventually preexisting old token
                localStorage.setItem('emailConfirmToken', this.$route.query.userEmailConfirm)
                router.push({name: 'login'})
			}
		},
        created() {
            if (this.user.authenticated) {
                this.loadNamespaces()
			}
        },
        watch: {
            // call the method again if the route changes
            '$route': 'doStuffAfterRoute'
        },
        methods: {
            logout() {
                auth.logout()
            },
			gravatar() {
                return 'https://www.gravatar.com/avatar/' + this.user.infos.avatar + '?s=50'
			},
            loadNamespaces() {
				const cancel = message.setLoading(this)
                HTTP.get(`namespaces`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
						this.$set(this, 'namespaces', response.data)
                        cancel()
                    })
                    .catch(e => {
						cancel()
                        this.handleError(e)
                    })
            },
			loadNamespacesIfNeeded(e){
                if (this.user.authenticated && e.name === 'home') {
                    this.loadNamespaces()
                }
			},
			doStuffAfterRoute(e) {
				this.fullpage = false;
				this.loadNamespacesIfNeeded(e)
				this.mobileMenuActive = false
			},
			setFullPage() {
				this.fullpage = true;
			},
            handleError(e) {
                message.error(e, this)
            }
        },
    }
</script>
