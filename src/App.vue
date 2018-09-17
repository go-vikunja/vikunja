<template>
	<div id="app" class="container">
		<nav class="navbar" role="navigation" aria-label="main navigation" v-if="user.authenticated">
			<div class="navbar-brand">
				<div class="navbar-item logo">
					<img src="images/logo-full.svg"/>
				</div>
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
										<router-link :to="{name: 'editNamespace', params: {id: n.id} }" class="button nsettings">
												<span class="icon">
													<icon icon="cog"/>
												</span>
										</router-link>
										<router-link :to="{ name: 'newList', params: { id: n.id} }" class="button is-success nsettings" :key="n.id + 'newList'">
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
				<router-view/>
			</div>
		</div>
		<notifications position="bottom left" />
	</div>
</template>

<script>
    import auth from './auth'
    import {HTTP} from './http-common'
	import message from './message'

    export default {
        name: 'app',

        data() {
            return {
                user: auth.user,
                loading: false,
                namespaces: [],
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

<style lang="scss">
	/* Logout-icon */
	.logout-icon {
		padding-right: 2em !important;
	}

	/* Logo */
	.logo {

		padding-left: 2rem !important;

		img {
			max-height: 3rem !important;
			margin-right: 1rem;
		}
	}

	/* Buttons icons */
	.button .icon.is-small {
		margin-right: 0.05rem !important;
	}

	/* List active link */
	.menu-list a.router-link-active{
		background: darken(#fff, 5%);
	}

	/* menu buttons */
	.button-bottom {
		margin-bottom: 1rem;
	}

	/* Namespaces list */
	.namespaces-lists{
		.menu-label {
			font-size: 1em;
			font-weight: 400;
			min-height: 2.5em;
			padding-top: 0.3em;
		}

		/* Namespace settings */
		.button{
			vertical-align: middle;
			float: right;
			margin-left: 0.5rem;
			min-width: 2.648em;
		}
	}
</style>
