<template>
	<div id="app" class="container">
		<nav class="navbar" role="navigation" aria-label="main navigation" v-if="user.authenticated">
			<div class="navbar-brand">
				<router-link :to="{name: 'home'}" class="navbar-item logo">
					<img src="images/logo-full.svg"/>
				</router-link>
			</div>
			<div class="navbar-menu">
				<div class="navbar-end">
					<span class="navbar-item">{{user.infos.username}}</span>
					<span class="navbar-item image">
						<img :src="gravatar()" class="is-rounded" alt=""/>
					</span>
					<a v-on:click="logout()" class="navbar-item is-right logout-icon">
						<span class="icon is-medium">
							<icon icon="sign-out-alt" size="2x"/>
						</span>
					</a>
				</div>
			</div>
		</nav>
		<div class="column is-centered">
			<div v-if="user.authenticated">
				<div class="box">
					<div class="columns">
						<div class="column is-3">
							<router-link :to="{name: 'listTeams'}" class="button is-primary is-fullwidth button-bottom">
								<span class="icon is-small">
									<icon icon="users"/>
								</span>
								Teams
							</router-link>
							<router-link :to="{name: 'newNamespace'}" class="button is-success is-fullwidth button-bottom">
								<span class="icon is-small">
									<icon icon="layer-group"/>
								</span>
								New Namespace
							</router-link>
							<aside class="menu namespaces-lists">
								<p class="menu-label" v-if="loading">Loading...</p>
								<template v-for="n in namespaces">
									<div :key="n.id">
										<router-link :to="{name: 'editNamespace', params: {id: n.id} }" class="nsettings">
												<span class="icon">
													<icon icon="cog"/>
												</span>
										</router-link>
										<router-link :to="{ name: 'newList', params: { id: n.id} }" class="is-success nsettings" :key="n.id + 'newList'">
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
						<div class="column is-9">
							<router-view/>
						</div>
					</div>
				</div>
			</div>
			<div v-else>
				<div class="container has-text-centered">
					<div class="column is-4 is-offset-4">
						<img src="images/logo-full.svg"/>
							<router-view/>
					</div>
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
            '$route': 'loadNamespacesIfNeeded'
        },
        methods: {
            logout() {
                auth.logout()
            },
			gravatar() {
                return 'https://www.gravatar.com/avatar/' + this.user.infos.avatar + '?s=50'
			},
            loadNamespaces() {
                this.loading = true
                this.namespaces = []
                HTTP.get(`namespaces`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
						this.$set(this, 'namespaces', response.data)
                        this.loading = false
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
			loadNamespacesIfNeeded(e){
                if (this.user.authenticated && e.name === 'home') {
                    this.loadNamespaces()
                }
			},
            handleError(e) {
                this.loading = false
                message.error(e, this)
            }
        },
    }
</script>
