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
					<p class="control is-expanded" v-bind:class="{ 'is-loading': loading}">
						<multiselect
								v-model="newUser"
								:options="foundUsers"
								:multiple="false"
								:searchable="true"
								:loading="loading"
								:internal-search="true"
								@search-change="findUsers"
								placeholder="Type to search a user"
								label="username"
								track-by="user_id">

							<template slot="clear" slot-scope="props">
								<div class="multiselect__clear" v-if="newUser.id !== 0" @mousedown.prevent.stop="clearAll(props.search)"></div>
							</template>
							<span slot="noResult">Oops! No users found. Consider changing the search query.</span>
						</multiselect>
					</p>
					<p class="control">
						<button type="submit" class="button is-success" style="margin-top: 3px;">
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
	import multiselect from 'vue-multiselect'
	import 'vue-multiselect/dist/vue-multiselect.min.css'

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
				newUser: {username: '', user_id: 0},
				userToDelete: 0,
				newUserid: 0,
				foundUsers: [],
			}
		},
		components: {
			multiselect
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
				const cancel = message.setLoading(this)
				HTTP.get(this.typeString + `s/` + this.id + `/users`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						//response.data.push(this.list.owner)
						this.$set(this, 'users', response.data)
						cancel()
					})
					.catch(e => {
						this.handleError(e)
						cancel()
					})
			},
			deleteUser() {
				const cancel = message.setLoading(this)
				HTTP.delete(this.typeString + `s/` + this.id + `/users/` + this.userToDelete, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.showUserDeleteModal = false;
						this.handleSuccess({message: 'The user was successfully deleted from the ' + this.typeString + '.'})
						this.loadUsers()
						cancel()
					})
					.catch(e => {
						this.handleError(e)
						cancel()
					})
			},
			addUser(admin) {
				const cancel = message.setLoading(this)
				if(admin === null) {
					admin = false
				}
				this.newUser.right = 0
				if (admin) {
					this.newUser.right = 2
				}

				this.$set(this, 'foundUsers', [])

				HTTP.put(this.typeString + `s/` + this.id + `/users`, this.newUser, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.loadUsers()
						this.newUser = {}
						this.handleSuccess({message: 'The user was successfully added.'})
						cancel()
					})
					.catch(e => {
						this.handleError(e)
						cancel()
					})
			},
			toggleUserType(userid, current) {
				const cancel = message.setLoading(this)
				let right = 0
				if (!current) {
					right = 2
				}

				HTTP.post(this.typeString + `s/` + this.id + `/users/` + userid, {right: right}, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.loadUsers()
						this.handleSuccess({message: 'The user right was successfully updated.'})
						cancel()
					})
					.catch(e => {
						this.handleError(e)
						cancel()
					})
			},
			findUsers(query) {
				const cancel = message.setLoading(this)
				if(query === '') {
					this.$set(this, 'foundUsers', [])
					cancel()
					return
				}

				this.$set(this, 'newUser', {username: '', user_id: 0})

				HTTP.get(`users?s=` + query, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						this.$set(this, 'foundUsers', [])

						for (const u in response.data) {
							this.foundUsers.push({
								username: response.data[u].username,
								user_id: response.data[u].id,
							})
						}

						cancel()
					})
					.catch(e => {
						this.handleError(e)
						cancel()
					})
			},
			clearAll () {
				this.$set(this, 'foundUsers', [])
			},
			limitText (count) {
				return `and ${count} others`
			},
			handleError(e) {
				message.error(e, this)
			},
			handleSuccess(e) {
				message.success(e, this)
			}
		},
	}
</script>

<style lang="scss">
	.card{
		margin-bottom: 1rem;

		.add-user-form {
			margin: 1rem;
		}

		.table{
			border-top: 1px solid darken(#fff, 15%);

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

	.users-list, .users-namespace{
		padding: 0 !important;
	}

	ul.multiselect__content{
		margin: 0 !important;
	}

	.multiselect{
		background-color: white;
		border: 1px solid #dbdbdb;
		color: #363636;
		-webkit-box-shadow: inset 0 1px 2px rgba(10, 10, 10, 0.1);
		box-shadow: inset 0 1px 2px rgba(10, 10, 10, 0.1);

		border-radius: 4px;
		font-size: 1rem;
		height: 2.25em;
		line-height: 1.5;
	}

	.multiselect--active{
		-webkit-box-shadow: inset 0 0.125em 0 rgba(10, 10, 10, 0.075), 0 0 0 0.125em rgba(91, 183, 219, 0.25);
		box-shadow: inset 0 0.125em 0 rgba(10, 10, 10, 0.075), 0 0 0 0.125em rgba(91, 183, 219, 0.25);
		border: 1px solid #5bb7db;
	}
</style>