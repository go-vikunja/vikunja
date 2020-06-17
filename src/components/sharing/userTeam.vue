<template>
	<div class="card is-fullwidth">
		<header class="card-header">
			<p class="card-header-title">
				Shared with these {{shareType}}s
			</p>
		</header>
		<div class="card-content content sharables-list">
			<form @submit.prevent="add()" class="add-form" v-if="userIsAdmin">
				<div class="field is-grouped">
					<p class="control is-expanded" v-bind:class="{ 'is-loading': searchService.loading}">
						<multiselect
								v-model="sharable"
								:options="found"
								:multiple="false"
								:searchable="true"
								:loading="searchService.loading"
								:internal-search="true"
								@search-change="find"
								placeholder="Type to search"
								:label="searchLabel"
								track-by="id">
							<template slot="clear" slot-scope="props">
								<div class="multiselect__clear" v-if="sharable.id !== 0" @mousedown.prevent.stop="clearAll(props.search)"></div>
							</template>
							<span slot="noResult">Oops! No {{shareType}} found. Consider changing the search query.</span>
						</multiselect>
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
				<tr v-for="s in sharables" :key="s.id">
					<template v-if="shareType === 'user'">
						<td>{{s.username}}</td>
						<td>
							<template v-if="s.id === userInfo.id">
								<b class="is-success">You</b>
							</template>
						</td>
					</template>
					<template v-if="shareType === 'team'">
						<td>
							<router-link :to="{name: 'teams.edit', params: {id: s.id}}">
								{{s.name}}
							</router-link>
						</td>
					</template>
					<td class="type">
						<template v-if="s.right === rights.ADMIN">
							<span class="icon is-small">
								<icon icon="lock"/>
							</span>
							Admin
						</template>
						<template v-else-if="s.right === rights.READ_WRITE">
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
						<div class="select">
							<select @change="toggleType(s)" v-model="selectedRight[s.id]" class="button buttonright">
								<option :value="rights.READ" :selected="s.right === rights.READ">Read only</option>
								<option :value="rights.READ_WRITE" :selected="s.right === rights.READ_WRITE">Read & write</option>
								<option :value="rights.ADMIN" :selected="s.right === rights.ADMIN">Admin</option>
							</select>
						</div>
						<button @click="() => {sharable = s; showDeleteModal = true}" class="button is-danger icon-only">
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
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				@submit="deleteSharable()">
			<span slot="header">Remove a {{shareType}} from the {{typeString}}</span>
			<p slot="text">Are you sure you want to remove this {{shareType}} from the {{typeString}}?<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
	import multiselect from 'vue-multiselect'
	import {mapState} from 'vuex'

	import UserNamespaceService from '../../services/userNamespace'
	import UserNamespaceModel from '../../models/userNamespace'
	import UserListModel from '../../models/userList'
	import UserListService from '../../services/userList'
	import UserService from '../../services/user'
	import UserModel from '../../models/user'

	import TeamNamespaceService from '../../services/teamNamespace'
	import TeamNamespaceModel from '../../models/teamNamespace'
	import TeamListModel from '../../models/teamList'
	import TeamListService from '../../services/teamList'
	import TeamService from '../../services/team'
	import TeamModel from '../../models/team'

	import rights from '../../models/rights'

	export default {
		name: 'userTeamShare',
		props: {
			type: {
				type: String,
				default: '',
			},
			shareType: {
				type: String,
				default: '',
			},
			id: {
				type: Number,
				default: 0,
			},
			userIsAdmin: {
				type: Boolean,
				default: false,
			},
		},
		data() {
			return {
				stuffService: Object, // This user service is either a userNamespaceService or a userListService, depending on the type we are using
				stuffModel: Object,
				searchService: Object,
				sharable: Object,

				found: [],
				searchLabel: '',
				rights: rights,
				selectedRight: {},

				typeString: '',
				sharables: [], // This holds either teams or users who this namepace or list is shared with
				showDeleteModal: false,
			}
		},
		components: {
			multiselect
		},
		computed: mapState({
			userInfo: state => state.auth.info
		}),
		created() {

			if (this.shareType === 'user') {
				this.searchService = new UserService()
				this.sharable = new UserModel()
				this.searchLabel = 'username'

				if (this.type === 'list') {
					this.typeString = `list`
					this.stuffService = new UserListService()
					this.stuffModel = new UserListModel({listId: this.id})
				} else if (this.type === 'namespace') {
					this.typeString = `namespace`
					this.stuffService = new UserNamespaceService()
					this.stuffModel = new UserNamespaceModel({namespaceId: this.id})
				} else {
					throw new Error('Unknown type: ' + this.type)
				}
			}
			else if (this.shareType === 'team') {
				this.searchService = new TeamService()
				this.sharable = new TeamModel()
				this.searchLabel = 'name'

				if (this.type === 'list') {
					this.typeString = `list`
					this.stuffService = new TeamListService()
					this.stuffModel = new TeamListModel({listId: this.id})
				} else if (this.type === 'namespace') {
					this.typeString = `namespace`
					this.stuffService = new TeamNamespaceService()
					this.stuffModel = new TeamNamespaceModel({namespaceId: this.id})
				} else {
					throw new Error('Unknown type: ' + this.type)
				}
			} else {
				throw new Error('Unkown share type')
			}

			this.load()
		},
		methods: {
			load() {
				this.stuffService.getAll(this.stuffModel)
					.then(r => {
						this.$set(this, 'sharables', r)
						r.forEach(s => this.$set(this.selectedRight, s.id, s.right))
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			deleteSharable() {

				if (this.shareType === 'user') {
					this.stuffModel.userId = this.sharable.username
				} else if (this.shareType === 'team') {
					this.stuffModel.teamId = this.sharable.id
				}
				this.stuffService.delete(this.stuffModel)
					.then(() => {
						this.showDeleteModal = false
						for (const i in this.sharables) {
							if (
								(this.sharables[i].id === this.stuffModel.userId && this.shareType === 'user') ||
								(this.sharables[i].id === this.stuffModel.teamId && this.shareType === 'team')
							) {
								this.sharables.splice(i, 1)
							}
						}
						this.success({message: 'The ' + this.shareType + ' was successfully deleted from the ' + this.typeString + '.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			add(admin) {
				if(admin === null) {
					admin = false
				}
				this.stuffModel.right = rights.READ
				if (admin) {
					this.stuffModel.right = rights.ADMIN
				}

				if (this.shareType === 'user') {
					this.stuffModel.userId = this.sharable.username
				} else if (this.shareType === 'team') {
					this.stuffModel.teamId = this.sharable.id
				}

				this.stuffService.create(this.stuffModel)
					.then(() => {
						this.success({message: 'The ' + this.shareType + ' was successfully added.'}, this)
						this.load()
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			toggleType(sharable) {
				if (this.selectedRight[sharable.id] !== rights.ADMIN &&
					this.selectedRight[sharable.id] !== rights.READ &&
					this.selectedRight[sharable.id] !== rights.READ_WRITE
				) {
					this.selectedRight[sharable.id] = rights.READ
				}
				this.stuffModel.right = this.selectedRight[sharable.id]


				if (this.shareType === 'user') {
					this.stuffModel.userId = sharable.username
				} else if (this.shareType === 'team') {
					this.stuffModel.teamId = sharable.id
				}

				this.stuffService.update(this.stuffModel)
					.then(r => {
						for (const i in this.sharables) {
							if (
								(this.sharables[i].username === this.stuffModel.userId && this.shareType === 'user') ||
								(this.sharables[i].id === this.stuffModel.teamId && this.shareType === 'team')
							) {
								this.$set(this.sharables[i], 'right', r.right)
							}
						}
						this.success({message: 'The ' + this.shareType + ' right was successfully updated.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			find(query) {
				if(query === '') {
					this.$set(this, 'found', [])
					return
				}

				this.searchService.getAll({}, {s: query})
					.then(response => {
						this.$set(this, 'found', response)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			clearAll () {
				this.$set(this, 'found', [])
			},
		},
	}
</script>
