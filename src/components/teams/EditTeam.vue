<template>
	<div>
		<div class="card" v-if="userIsAdmin">
			<header class="card-header">
				<p class="card-header-title">
					Edit Team
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form  @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="teamtext">Team Name</label>
							<div class="control">
								<input :class="{ 'disabled': loading}" :disabled="loading" class="input" type="text" id="teamtext" placeholder="The team text is here..." v-model="team.name">
							</div>
						</div>
						<div class="field">
							<label class="label" for="teamdescription">Description</label>
							<div class="control">
								<textarea :class="{ 'disabled': loading}" :disabled="loading" class="textarea" placeholder="The teams description goes here..." id="teamdescription" v-model="team.description"></textarea>
							</div>
						</div>
					</form>

					<div class="columns bigbuttons">
						<div class="column">
							<button @click="submit()" class="button is-success is-fullwidth" :class="{ 'is-loading': loading}">
								Save
							</button>
						</div>
						<div class="column is-1">
							<button @click="showDeleteModal = true" class="button is-danger is-fullwidth" :class="{ 'is-loading': loading}">
								<span class="icon is-small">
									<icon icon="trash-alt"/>
								</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
		<div class="card">

			<header class="card-header">
				<p class="card-header-title">
					Team Members
				</p>
			</header>
			<div class="card-content content">
				<table>
					<tr v-for="m in team.members" :key="m.id">
						<td>{{m.username}}</td>
						<td>
							<template v-if="m.id === user.infos.id">
								<b class="is-success">You</b>
							</template>
						</td>
						<td><template v-if="m.admin">Is Admin</template><template v-else>Normal Member</template></td>
					</tr>
				</table>
			</div>
		</div>

		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				v-on:submit="deleteTeam()">
			<span slot="header">Delete the team</span>
			<p slot="text">Are you sure you want to delete this team and all of its members?<br/>
				All team members will loose access to lists and namespaces shared with this team.<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'
    import {HTTP} from '../../http-common'
    import message from '../../message'
	
    export default {
        name: "EditTeam",
        data() {
            return {
                team: {title: '', description:''},
                error: '',
                loading: false,
                showDeleteModal: false,
				user: auth.user,
				userIsAdmin: false,
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        created() {
            this.loadTeam()
        },
        watch: {
            // call again the method if the route changes
            '$route': 'loadTeam'
        },
		methods: {
            loadTeam() {
                this.loading = true

                HTTP.get(`teams/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.$set(this, 'team', response.data)
						let members = response.data.members
						for (const m in members) {
							if (members[m].id === this.user.infos.id && members[m].admin) {
								this.userIsAdmin = true
							}
						}
                        this.loading = false
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            submit() {
                this.loading = true

                HTTP.post(`teams/` + this.$route.params.id, this.team, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        // Update the team in the parent
                        for (const n in this.$parent.teams) {
                            if (this.$parent.teams[n].id === response.data.id) {
                                response.data.lists = this.$parent.teams[n].lists
                                this.$set(this.$parent.teams, n, response.data)
                            }
                        }
                        this.handleSuccess({message: 'The team was successfully updated.'})
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            deleteTeam() {
                HTTP.delete(`teams/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(() => {
                        this.handleSuccess({message: 'The team was successfully deleted.'})
                        router.push({name: 'home'})
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
			},
            handleError(e) {
                this.loading = false
                message.error(e, this)
            },
            handleSuccess(e) {
                this.loading = false
                message.success(e, this)
            }
		}
    }
</script>

<style scoped>
	.bigbuttons{
		margin-top: 0.5rem;
	}

	.card{
		margin-bottom: 1rem;
	}
</style>