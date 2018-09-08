<template>
	<div id="app" class="container">
		<div class="column is-centered" v-if="user.authenticated">
			<button v-on:click="logout()" class="button is-right">Logout</button>
		</div>
		<div class="column is-centered">
			<div class="box" v-if="user.authenticated">
				<div class="container">
					<div class="columns">
						<div class="column is-3">
							<aside class="menu">
								<p class="menu-label" v-if="loading">Loading...</p>
								<template v-for="n in namespaces">
									<p class="menu-label" :key="n.id">
										{{n.name}}
									</p>
									<ul class="menu-list" :key="n.id + 'child'">
										<li v-for="l in n.lists" :key="l.id">
											<a>{{l.title}}</a>
										</li>
									</ul>
								</template>
							</aside>
						</div>
						<div class="column is-9">
							<button class="button is-success" v-on:click="loadNamespaces()">Load Namespaces</button>
							<router-view/>
						</div>
					</div>
				</div>
			</div>
			<div v-else>
				<router-view/>
			</div>
		</div>
	</div>
</template>

<script>
    import auth from './auth'
    import {HTTP} from './http-common'

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
        methods: {
            logout() {
                auth.logout()
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
                                    this.loading = false
                                    // eslint-disable-next-line
                                    console.log(e)
                                })
                        }

                        this.loading = false
                    })
                    .catch(e => {
                        this.loading = false
                        // eslint-disable-next-line
                        console.log(e)
                    })
            },
        },
    }
</script>

<style>
	body {
		background: #f5f5f5;
		min-height: 100vh;
	}
</style>
