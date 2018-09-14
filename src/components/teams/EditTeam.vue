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
			<div class="card-content content team-members">
				<table class="table is-striped is-hoverable is-fullwidth">
					<tbody>
						<tr v-for="m in team.members" :key="m.id">
							<td>{{m.username}}</td>
							<td>
								<template v-if="m.id === user.infos.id">
									<b class="is-success">You</b>
								</template>
							</td>
							<td class="type">
								<template v-if="m.admin">
									<span class="icon is-small">
										<icon icon="lock"/>
									</span>
									Admin
								</template>
								<template v-else>
									<span class="icon is-small">
										<icon icon="user"/>
									</span>
									Member
								</template>
							</td>
							<td class="actions" v-if="userIsAdmin">
								<button @click="toggleUserType(m.id, m.admin)" class="button buttonright is-primary" v-if="m.id !== user.infos.id">
									Make
									<template v-if="!m.admin">
										Admin
									</template>
									<template v-else>
										Member
									</template>
								</button>
								<button @click="userToDelete = m.id; showUserDeleteModal()" class="button is-danger" v-if="m.id !== user.infos.id">
									<span class="icon is-small">
										<icon icon="trash-alt"/>
									</span>
								</button>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>

		<!-- Team delete modal -->
		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				v-on:submit="deleteTeam()">
			<span slot="header">Delete the team</span>
			<p slot="text">Are you sure you want to delete this team and all of its members?<br/>
				All team members will loose access to lists and namespaces shared with this team.<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
		<!-- User delete modal -->
		<modal
				v-if="showUserDeleteModal"
				@close="showUserDeleteModal = false"
				v-on:submit="deleteUser(this.userToDelete)">
			<span slot="header">Remove a user from the team</span>
			<p slot="text">Are you sure you want to remove this user from the team?<br/>
				He will loose access to all lists and namespaces this team has access to.<br/>
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
				showUserDeleteModal: false,
				user: auth.user,
				userIsAdmin: false,
				userToDelete: 0,
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
			deleteUser() {
				HTTP.delete(`teams/` + this.$route.params.id + `/members/` + this.userToDelete, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.handleSuccess({message: 'The user was successfully deleted from the team.'})
						this.loadTeam()
					})
					.catch(e => {
						this.handleError(e)
					})
			},
			addUser(userid, admin) {
				if(admin === null) {
					admin = false
				}
				HTTP.put(`teams/` + this.$route.params.id + `/members`, {admin: admin, user_id: userid}, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.handleSuccess({message: 'The team was successfully added.'})
					})
					.catch(e => {
						this.handleError(e)
					})
			},
			toggleUserType(userid, current) {
				this.userToDelete = userid
				this.deleteUser()
				this.addUser(userid, !current)
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

<style lang="scss" scoped>
	.bigbuttons{
		margin-top: 0.5rem;
	}

	.card{
		margin-bottom: 1rem;

		.table{
			td{
				vertical-align: middle;
			}

			td.type, td.actions{
				width: 200px;
			}

			td.actions{
				text-align: right;
			}
		}
	}

	.buttonright {
		margin-right: 0.5rem;
	}

	.team-members{
		padding: 0;
	}
</style>