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
								<template v-for="n in namespaces">
									<p v-bind:key="n.id" class="menu-label">
										{{n.name}}
									</p>
									<ul v-bind:key="n.lists" class="menu-list">
										<template v-for="l in n.lists">
											<li v-bind:key="l.id"><a>{{l.title}}</a></li>
										</template>
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
            this.loadNamespaces()
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

                        let namespaces = response.data

                        // Get the lists
                        for (const n in namespaces) {

                            this.namespaces[n] = namespaces[n]

                            HTTP.get(`namespaces/` + namespaces[n].id + `/lists`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                                .then(response => {
                                    this.namespaces[n].lists = response.data
                                })
                                .catch(e => {
                                    this.loading = false
                                    // eslint-disable-next-line
                                    console.log(e)
                                })
                        }

                        // eslint-disable-next-line
                        console.log(this.namespaces)

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
