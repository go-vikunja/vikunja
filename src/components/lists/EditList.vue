<template>
	<div>
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Edit List
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form  @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="listtext">List Name</label>
							<div class="control">
								<input :class="{ 'disabled': loading}" :disabled="loading" class="input" type="text" id="listtext" placeholder="The list title goes here..." v-model="list.title">
							</div>
						</div>
						<div class="field">
							<label class="label" for="listdescription">Description</label>
							<div class="control">
								<textarea :class="{ 'disabled': loading}" :disabled="loading" class="textarea" placeholder="The lists description goes here..." id="listdescription" v-model="list.description"></textarea>
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
					Users with access to this list
				</p>
			</header>
			<div class="card-content content users-list">
				<form @submit.prevent="addUser()" class="add-user-form" v-if="userIsAdmin">
					<div class="field is-grouped">
						<p class="control has-icons-left is-expanded" v-bind:class="{ 'is-loading': loading}">
							<input class="input" v-bind:class="{ 'disabled': loading}" v-model.number="newUser.user_id" type="text" placeholder="Add a new user...">
							<span class="icon is-small is-left">
								<icon icon="user"/>
							</span>
						</p>
						<p class="control">
							<button type="submit" class="button is-success">
								<span class="icon is-small">
									<icon icon="plus"/>
								</span>
								Add
							</button>
						</p>
					</div>
				</form>
				<table class="table is-striped is-hoverable is-fullwidth">
					<tbody>
					<tr v-for="u in listUsers" :key="u.id">
						<td>{{u.username}}</td>
						<td>
							<template v-if="u.id === user.infos.id">
								<b class="is-success">You</b>
							</template>
						</td>
						<td class="type">
							<template v-if="u.right === 2">
									<span class="icon is-small">
										<icon icon="lock"/>
									</span>
								Admin
							</template>
							<template v-else-if="u.right === 1">
									<span class="icon is-small">
										<icon icon="pen"/>
									</span>
								Write
							</template>
							<template v-else>
									<span class="icon is-small">
										<icon icon="users"/>
									</span>
								Read-only
							</template>
						</td>
						<td class="actions" v-if="userIsAdmin">
							<button @click="toggleUserType(u.id, (u.right === 2))" class="button buttonright is-primary" v-if="u.id !== user.infos.id">
								Make
								<template v-if="u.right === 2">
									Member
								</template>
								<template v-else>
									Admin
								</template>
							</button>
							<button @click="userToDelete = u.id; showUserDeleteModal = true" class="button is-danger" v-if="u.id !== user.infos.id">
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

		<div class="card">

			<header class="card-header">
				<p class="card-header-title">
					Teams with access to this list
				</p>
			</header>
			<div class="card-content content teams-list">
				<form @submit.prevent="addTeam()" class="add-team-form" v-if="userIsAdmin">
					<div class="field is-grouped">
						<p class="control has-icons-left is-expanded" v-bind:class="{ 'is-loading': loading}">
							<input class="input" v-bind:class="{ 'disabled': loading}" v-model.number="newTeam.team_id" type="text" placeholder="Add a new team...">
							<span class="icon is-small is-left">
								<icon icon="users"/>
							</span>
						</p>
						<p class="control">
							<button type="submit" class="button is-success">
								<span class="icon is-small">
									<icon icon="plus"/>
								</span>
								Add
							</button>
						</p>
					</div>
				</form>
				<table class="table is-striped is-hoverable is-fullwidth">
					<tbody>
					<tr v-for="t in listTeams" :key="t.id">
						<td>
							<router-link :to="{name: 'editTeam', params: {id: t.id}}">
								{{t.name}}
							</router-link>
						</td>
						<td class="type">
							<template v-if="t.right === 2">
									<span class="icon is-small">
										<icon icon="lock"/>
									</span>
								Admin
							</template>
							<template v-else-if="t.right === 1">
									<span class="icon is-small">
										<icon icon="pen"/>
									</span>
								Write
							</template>
							<template v-else>
									<span class="icon is-small">
										<icon icon="users"/>
									</span>
								Read-only
							</template>
						</td>
						<td class="actions" v-if="userIsAdmin">
							<button @click="toggleTeamType(t.id, (t.right === 2))" class="button buttonright is-primary">
								Make
								<template v-if="t.right === 2">
									Member
								</template>
								<template v-else>
									Admin
								</template>
							</button>
							<button @click="teamToDelete = t.id; showTeamDeleteModal = true" class="button is-danger">
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
		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				v-on:submit="deleteList()">
			<span slot="header">Delete the list</span>
			<p slot="text">Are you sure you want to delete this list and all of its contents?
				<br/>This includes all tasks and <b>CANNOT BE UNDONE!</b></p>
		</modal>

		<!-- User delete modal -->
		<modal
				v-if="showUserDeleteModal"
				@close="showUserDeleteModal = false"
				v-on:submit="deleteUser()">
			<span slot="header">Remove a user from the list</span>
			<p slot="text">Are you sure you want to remove this user from the list?<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>

		<!-- Team delete modal -->
		<modal
				v-if="showTeamDeleteModal"
				@close="showTeamDeleteModal = false"
				v-on:submit="deleteTeam()">
			<span slot="header">Remove a team from the list</span>
			<p slot="text">Are you sure you want to remove this team from the list?<br/>
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
        name: "EditList",
        data() {
            return {
                list: {title: '', description:''},
                error: '',
                loading: false,
                showDeleteModal: false,
				user: auth.user,
				userIsAdmin: false,

				listUsers: [],
				newUser: {user_id: 0},
				showUserDeleteModal: false,
				userToDelete: 0,

				listTeams: [],
				newTeam: {team_id: 0},
				showTeamDeleteModal: false,
				teamToDelete: 0,
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        created() {
            this.loadList()
        },
        watch: {
            // call again the method if the route changes
            '$route': 'loadList'
        },
        methods: {
            loadList() {
                this.loading = true

                HTTP.get(`lists/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.$set(this, 'list', response.data)
						if (response.data.owner.id === this.user.infos.id) {
							this.userIsAdmin = true
						}
						this.loadUsers()
						this.loadTeams()
                        this.loading = false
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            submit() {
                this.loading = true

                HTTP.post(`lists/` + this.$route.params.id, this.list, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        // Update the list in the parent
                        for (const n in this.$parent.namespaces) {
                            let lists = this.$parent.namespaces[n].lists
                            for (const l in lists) {
                                if (lists[l].id === response.data.id) {
                                    this.$set(this.$parent.namespaces[n].lists, l, response.data)
                                }
                            }
                        }
                        this.handleSuccess({message: 'The list was successfully updated.'})
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            deleteList() {
                HTTP.delete(`lists/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(() => {
                        this.handleSuccess({message: 'The list was successfully deleted.'})
                        router.push({name: 'home'})
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
			loadUsers() {
				// Get all users with access to the list
				HTTP.get(`lists/` + this.$route.params.id + `/users`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						response.data.push(this.list.owner)
						this.$set(this, 'listUsers', response.data)
						this.loading = false
					})
					.catch(e => {
						this.handleError(e)
					})
			},
			deleteUser() {
				HTTP.delete(`lists/` + this.$route.params.id + `/users/` + this.userToDelete, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.showUserDeleteModal = false;
						this.handleSuccess({message: 'The user was successfully deleted from the list.'})
						this.loadUsers()
					})
					.catch(e => {
						this.handleError(e)
					})
			},
			addUser(admin) {
				if(admin === null) {
					admin = false
				}
				this.newUser.right = 0
				if (admin) {
					this.newUser.right = 2
				}

				HTTP.put(`lists/` + this.$route.params.id + `/users`, this.newUser, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.loadUsers()
						this.handleSuccess({message: 'The user was successfully added.'})
					})
					.catch(e => {
						this.handleError(e)
					})
			},
			toggleUserType(userid, current) {
				this.userToDelete = userid
				this.newUser.user_id = userid
				this.deleteUser()
				this.addUser(!current)
			},
			loadTeams() {
				HTTP.get(`lists/` + this.$route.params.id + `/teams`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						this.$set(this, 'listTeams', response.data)
						this.loading = false
					})
					.catch(e => {
						this.handleError(e)
					})
			},
			deleteTeam() {
				HTTP.delete(`lists/` + this.$route.params.id + `/teams/` + this.teamToDelete, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.showTeamDeleteModal = false;
						this.handleSuccess({message: 'The team was successfully deleted from the list.'})
						this.loadTeams()
					})
					.catch(e => {
						this.handleError(e)
					})
			},
			addTeam(admin) {
				if(admin === null) {
					admin = false
				}
				this.newTeam.right = 0
				if (admin) {
					this.newTeam.right = 2
				}

				HTTP.put(`lists/` + this.$route.params.id + `/teams`, this.newTeam, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.loadTeams()
						this.handleSuccess({message: 'The team was successfully added.'})
					})
					.catch(e => {
						this.handleError(e)
					})
			},
			toggleTeamType(teamid, current) {
				this.teamToDelete = teamid
				this.newTeam.team_id = teamid
				this.deleteTeam()
				this.addTeam(!current)
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
