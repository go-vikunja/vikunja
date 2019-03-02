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
					<p class="control is-expanded" v-bind:class="{ 'is-loading': userStuffService.loading}">
						<multiselect
								v-model="user"
								:options="foundUsers"
								:multiple="false"
								:searchable="true"
								:loading="userService.loading"
								:internal-search="true"
								@search-change="findUsers"
								placeholder="Type to search a user"
								label="username"
								track-by="id">
							<template slot="clear" slot-scope="props">
								<div class="multiselect__clear" v-if="user.id !== 0" @mousedown.prevent.stop="clearAll(props.search)"></div>
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
						<button @click="user = u; showUserDeleteModal = true" class="button is-danger" v-if="u.id !== currentUser.id">
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
	import auth from '../../auth'
	import message from '../../message'
	import multiselect from 'vue-multiselect'
	import 'vue-multiselect/dist/vue-multiselect.min.css'

	import UserService from '../../services/user'
	import UserNamespaceModel from '../../models/userNamespace'
	import UserListModel from '../../models/userList'
	import UserListService from '../../services/userList'
	import UserNamespaceService from '../../services/userNamespace'
	import UserModel from '../../models/user'

	export default {
		name: 'user',
		props: {
			type: '',
			id: 0,
			userIsAdmin: false,
		},
		data() {
			return {
				userService: UserService, // To search for users
				user: UserModel,
				userStuff: Object, // This will be either UserNamespaceModel or UserListModel
				userStuffService: Object, // This will be either UserListService or UserNamespaceService

				currentUser: auth.user.infos,
				typeString: '',
				showUserDeleteModal: false,
				users: [],
				foundUsers: [],
			}
		},
		components: {
			multiselect
		},
		created() {
			this.userService = new UserService()
			this.user = new UserModel()

			if (this.type === 'list') {
				this.typeString = `list`
				this.userStuffService = new UserListService()
				this.userStuff = new UserListModel({listID: this.id})
			} else if (this.type === 'namespace') {
				this.typeString = `namespace`
				this.userStuffService = new UserNamespaceService()
				this.userStuff = new UserNamespaceModel({namespaceID: this.id})
			} else {
				throw new Error('Unknown type: ' + this.type)
			}

			this.loadUsers()
		},
		methods: {
			loadUsers() {
				this.userStuffService.getAll(this.userStuff)
					.then(response => {
						this.$set(this, 'users', response)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			deleteUser() {
				// The api wants the user id as userID
				let usr = this.user
				this.userStuff.userID = usr.id

				this.userStuffService.delete(this.userStuff)
					.then(() => {
						this.showUserDeleteModal = false;
						message.success({message: 'The user was successfully deleted from the ' + this.typeString + '.'}, this)
						this.loadUsers()
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			addUser(admin = false) {
				this.userStuff.right = 0
				if (admin) {
					this.userStuff.right = 2
				}

				// The api wants the user id as userID
				this.userStuff.userID = this.user.id
				this.$set(this, 'foundUsers', [])

				this.userStuffService.create(this.userStuff)
					.then(() => {
						this.loadUsers()
						message.success({message: 'The user was successfully added.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			toggleUserType(userid, current) {
				this.userStuff.userID = userid
				this.userStuff.right = 0
				if (!current) {
					this.userStuff.right = 2
				}

				this.userStuffService.update(this.userStuff)
					.then(() => {
						this.loadUsers()
						message.success({message: 'The user right was successfully updated.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			findUsers(query) {
				if(query === '') {
					this.$set(this, 'foundUsers', [])
					return
				}

				this.userService.getAll({}, {s: query})
					.then(response => {
						this.$set(this, 'foundUsers', response)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			clearAll () {
				this.$set(this, 'foundUsers', [])
			},
			limitText (count) {
				return `and ${count} others`
			},
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