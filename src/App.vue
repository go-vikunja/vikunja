<template>
	<div id="app" class="container">
		<nav class="navbar" role="navigation" aria-label="main navigation" v-if="user.authenticated">
			<div class="navbar-menu">
				<div class="navbar-brand">
					<div class="navbar-item logo">
						<img src="logo-full.svg"/>
					</div>
				</div>

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
			<div class="box" v-if="user.authenticated">
				<div class="container">
					<div class="columns">
						<div class="column is-3">
							<div class="buttons">
								<a class="button is-success">
									<span class="icon is-small">
										<icon icon="plus"/>
									</span>
									New Namespace
								</a>
								<a class="button is-success">
									<span class="icon is-small">
										<icon icon="list-ol"/>
									</span>
									New List
								</a>
							</div>
							<aside class="menu">
								<p class="menu-label" v-if="loading">Loading...</p>
								<template v-for="n in namespaces">
									<p class="menu-label" :key="n.id">
										{{n.name}}
									</p>
									<ul class="menu-list" :key="n.id + 'child'">
										<li v-for="l in n.lists" :key="l.id">
											<router-link :to="{ name: 'showList', params: { id: l.id} }">{{l.title}}</router-link>
										</li>
									</ul>
								</template>
							</aside>
						</div>
						<div class="column">
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

                        let nps = response.data

                        // Loop through the namespaces and get their lists
                        for (const n in nps) {

                            this.namespaces.push(nps[n])

                            HTTP.get(`namespaces/` + nps[n].id + `/lists`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                                .then(response => {
                                    // This adds a new element "list" to our object which contains all lists
                                    this.$set(this.namespaces[n], 'lists', response.data)
                                })
                                .catch(e => {
                                    this.handleError(e)
                                })
                        }

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
	@import url('/fonts/fonts.css');

	*, *:focus, *:active{
		outline: none;
	}

	body {
		background: #f5f5f5;
		min-height: 100vh;
	}

	h1,h2,h3,h4,h5,h6{
		font-family: 'Quicksand', sans-serif;
		font-weight: 400 !important;
	}

	/* spinner */
	.fullscreen-loader-wrapper {
		position: fixed;
		background: rgba(250,250,250,0.8);
		z-index: 5;
		top: 0;
		bottom: 0;
		left: 0;
		right: 0;

		position: absolute;
		width: 78%;
		height: 100%;
		margin: -1em auto;
	}

	.full-loader-wrapper{
		background: rgba(250,250,250,0.8);

		position: absolute;
		width: 78%;
		height: 100%;
		margin: -1em auto;
	}

	.half-circle-spinner, .half-circle-spinner * {
		box-sizing: border-box;
	}

	.half-circle-spinner {
		width: 60px;
		height: 60px;
		border-radius: 100%;
		position: relative;
		left: calc(50% - 30px);
		top: calc(50% - 30px);
	}

	.half-circle-spinner .circle {
		content: "";
		position: absolute;
		width: 100%;
		height: 100%;
		border-radius: 100%;
		border: calc(60px / 10) solid transparent;
	}

	.half-circle-spinner .circle.circle-1 {
		border-top-color: #4CAF50;
		animation: half-circle-spinner-animation 1s infinite;
	}

	.half-circle-spinner .circle.circle-2 {
		border-bottom-color: #4CAF50;
		animation: half-circle-spinner-animation 1s infinite alternate;
	}

	@keyframes half-circle-spinner-animation {
		0% {
			transform: rotate(0deg);

		}
		100%{
			transform: rotate(360deg);
		}
	}

	/* Logout-icon */
	.logout-icon {
		padding-right: 2em !important;
	}

	/* Navbar */
	.navbar{
		background: none !important;
		padding-top: 0.5rem;
	}

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
</style>
