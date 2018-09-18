<template>
	<div class="card">

		<header class="card-header">
			<p class="card-header-title">
				Users with access to this {{type}}
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
				<tr v-for="u in users" :key="u.id">
					<td>{{u.username}}</td>
					<td>
						<template v-if="u.id === currentUser.id">
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
						<button @click="toggleUserType(u.id, (u.right === 2))" class="button buttonright is-primary" v-if="u.id !== currentUser.id">
							Make
							<template v-if="u.right === 2">
								Member
							</template>
							<template v-else>
								Admin
							</template>
						</button>
						<button @click="userToDelete = u.id; showUserDeleteModal = true" class="button is-danger" v-if="u.id !== currentUser.id">
									<span class="icon is-small">
										<icon icon="trash-alt"/>
									</span>
						</button>
					</td>
				</tr>
				</tbody>
			</table>
		</div>

		<modal
				v-if="showUserDeleteModal"
				@close="showUserDeleteModal = false"
				v-on:submit="deleteUser()">
			<span slot="header">Remove a user from the {{typeString}}</span>
			<p slot="text">Are you sure you want to remove this user from the {{typeString}}?<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
	import {HTTP} from '../../http-common'
	import auth from '../../auth'
	import message from '../../message'

	export default {
		name: 'user',
		props: {
			type: '',
			id: 0,
			userIsAdmin: false,
		},
		data() {
			return {
				loading: false,
				currentUser: auth.user.infos,
				typeString: '',
				showUserDeleteModal: false,
				users: [],
				newUser: {user_id: 0},
				userToDelete: 0,
			}
		},
		created() {
			if (this.type === 'list') {
				this.typeString = `list`
			} else if (this.type === 'namespace') {
				this.typeString = `namespace`
			} else {
				throw new Error('Unknown type: ' + this.type)
			}

			this.loadUsers()
		},
		methods: {
			loadUsers() {
				HTTP.get(this.typeString + `s/` + this.id + `/users`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						//response.data.push(this.list.owner)
						this.$set(this, 'users', response.data)
						this.loading = false
					})
					.catch(e => {
						this.handleError(e)
					})
			},
			deleteUser() {
				HTTP.delete(this.typeString + `s/` + this.id + `/users/` + this.userToDelete, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.showUserDeleteModal = false;
						this.handleSuccess({message: 'The user was successfully deleted from the ' + this.typeString + '.'})
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

				HTTP.put(this.typeString + `s/` + this.id + `/users`, this.newUser, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
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
			handleError(e) {
				this.loading = false
				message.error(e, this)
			},
			handleSuccess(e) {
				this.loading = false
				message.success(e, this)
			}
		},
	}
</script>

<style scoped>

</style>