<template>
	<div>
		<p class="has-text-weight-bold">
			{{ $t('list.share.userTeam.shared', {type: shareTypeNames}) }}
		</p>
		<div v-if="userIsAdmin">
			<div class="field has-addons">
				<p
					class="control is-expanded"
					:class="{ 'is-loading': searchService.loading }"
				>
					<multiselect
						:loading="searchService.loading"
						:placeholder="$t('misc.searchPlaceholder')"
						@search="find"
						:search-results="found"
						:label="searchLabel"
						v-model="sharable"
					/>
				</p>
				<p class="control">
					<x-button @click="add()">{{ $t('list.share.share') }}</x-button>
				</p>
			</div>
		</div>

		<table class="table has-actions is-striped is-hoverable is-fullwidth mb-4" v-if="sharables.length > 0">
			<tbody>
			<tr :key="s.id" v-for="s in sharables">
				<template v-if="shareType === 'user'">
					<td>{{ s.getDisplayName() }}</td>
					<td>
						<template v-if="s.id === userInfo.id">
							<b class="is-success">{{ $t('list.share.userTeam.you') }}</b>
						</template>
					</td>
				</template>
				<template v-if="shareType === 'team'">
					<td>
						<router-link
							:to="{
									name: 'teams.edit',
									params: { id: s.id },
								}"
						>
							{{ s.name }}
						</router-link>
					</td>
				</template>
				<td class="type">
					<template v-if="s.right === rights.ADMIN">
							<span class="icon is-small">
								<icon icon="lock"/>
							</span>
						{{ $t('list.share.right.admin') }}
					</template>
					<template v-else-if="s.right === rights.READ_WRITE">
							<span class="icon is-small">
								<icon icon="pen"/>
							</span>
						{{ $t('list.share.right.readWrite') }}
					</template>
					<template v-else>
							<span class="icon is-small">
								<icon icon="users"/>
							</span>
						{{ $t('list.share.right.read') }}
					</template>
				</td>
				<td class="actions" v-if="userIsAdmin">
					<div class="select">
						<select
							@change="toggleType(s)"
							class="button mr-2"
							v-model="selectedRight[s.id]"
						>
							<option
								:selected="s.right === rights.READ"
								:value="rights.READ"
							>
								{{ $t('list.share.right.read') }}
							</option>
							<option
								:selected="s.right === rights.READ_WRITE"
								:value="rights.READ_WRITE"
							>
								{{ $t('list.share.right.readWrite') }}
							</option>
							<option
								:selected="s.right === rights.ADMIN"
								:value="rights.ADMIN"
							>
								{{ $t('list.share.right.admin') }}
							</option>
						</select>
					</div>
					<x-button
						@click="
								() => {
									sharable = s
									showDeleteModal = true
								}
							"
						class="is-danger"
						icon="trash-alt"
					/>
				</td>
			</tr>
			</tbody>
		</table>

		<nothing v-else>
			{{ $t('list.share.userTeam.notShared', {type: shareTypeNames}) }}
		</nothing>

		<transition name="modal">
			<modal
				@close="showDeleteModal = false"
				@submit="deleteSharable()"
				v-if="showDeleteModal"
			>
				<template #header>
					<span>{{ $t('list.share.userTeam.removeHeader', {type: shareTypeName, sharable: sharableName}) }}</span>
				</template>
				<template #text>
					<p>{{ $t('list.share.userTeam.removeText', {type: shareTypeName, sharable: sharableName}) }}</p>
				</template>
			</modal>
		</transition>
	</div>
</template>

<script>
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

import rights from '../../models/constants/rights.json'
import Multiselect from '@/components/input/multiselect.vue'
import Nothing from '@/components/misc/nothing.vue'

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
		Nothing,
		Multiselect,
	},
	computed: {
		userInfo() {
			return this.$store.state.auth.info
		},
		shareTypeNames() {
			if (this.shareType === 'user') {
				return this.$tc('list.share.userTeam.typeUser', 2)
			}

			if (this.shareType === 'team') {
				return this.$tc('list.share.userTeam.typeTeam', 2)
			}

			return ''
		},
		shareTypeName() {
			if (this.shareType === 'user') {
				return this.$tc('list.share.userTeam.typeUser', 1)
			}

			if (this.shareType === 'team') {
				return this.$tc('list.share.userTeam.typeTeam', 1)
			}

			return ''
		},
		sharableName() {
			if (this.type === 'list') {
				return this.$t('list.list.title')
			}

			if (this.shareType === 'namespace') {
				return this.$t('namespace.namespace')
			}

			return ''
		},
	},
	created() {
		if (this.shareType === 'user') {
			this.searchService = new UserService()
			this.sharable = new UserModel()
			this.searchLabel = 'username'

			if (this.type === 'list') {
				this.typeString = 'list'
				this.stuffService = new UserListService()
				this.stuffModel = new UserListModel({listId: this.id})
			} else if (this.type === 'namespace') {
				this.typeString = 'namespace'
				this.stuffService = new UserNamespaceService()
				this.stuffModel = new UserNamespaceModel({
					namespaceId: this.id,
				})
			} else {
				throw new Error('Unknown type: ' + this.type)
			}
		} else if (this.shareType === 'team') {
			this.searchService = new TeamService()
			this.sharable = new TeamModel()
			this.searchLabel = 'name'

			if (this.type === 'list') {
				this.typeString = 'list'
				this.stuffService = new TeamListService()
				this.stuffModel = new TeamListModel({listId: this.id})
			} else if (this.type === 'namespace') {
				this.typeString = 'namespace'
				this.stuffService = new TeamNamespaceService()
				this.stuffModel = new TeamNamespaceModel({
					namespaceId: this.id,
				})
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
			this.stuffService
				.getAll(this.stuffModel)
				.then((r) => {
					this.sharables = r
					r.forEach((s) =>
						this.selectedRight[s.id] = s.right,
					)
				})
		},
		deleteSharable() {
			if (this.shareType === 'user') {
				this.stuffModel.userId = this.sharable.username
			} else if (this.shareType === 'team') {
				this.stuffModel.teamId = this.sharable.id
			}

			this.stuffService
				.delete(this.stuffModel)
				.then(() => {
					this.showDeleteModal = false
					for (const i in this.sharables) {
						if (
							(this.sharables[i].username === this.stuffModel.userId && this.shareType === 'user') ||
							(this.sharables[i].id === this.stuffModel.teamId && this.shareType === 'team')
						) {
							this.sharables.splice(i, 1)
						}
					}
					this.$message.success({message: this.$t('list.share.userTeam.removeSuccess', {type: this.shareTypeName, sharable: this.sharableName})})
				})
		},
		add(admin) {
			if (admin === null) {
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

			this.stuffService
				.create(this.stuffModel)
				.then(() => {
					this.$message.success({message: this.$t('list.share.userTeam.addedSuccess', {type: this.shareTypeName})})
					this.load()
				})
		},
		toggleType(sharable) {
			if (
				this.selectedRight[sharable.id] !== rights.ADMIN &&
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

			this.stuffService
				.update(this.stuffModel)
				.then((r) => {
					for (const i in this.sharables) {
						if (
							(this.sharables[i].username ===
								this.stuffModel.userId &&
								this.shareType === 'user') ||
							(this.sharables[i].id === this.stuffModel.teamId &&
								this.shareType === 'team')
						) {
							this.sharables[i].right = r.right
						}
					}
					this.$message.success({message: this.$t('list.share.userTeam.updatedSuccess', {type: this.shareTypeName})})
				})
		},
		find(query) {
			if (query === '') {
				this.clearAll()
				return
			}

			this.searchService
				.getAll({}, {s: query})
				.then((response) => {
					this.found = response
				})
		},
		clearAll() {
			this.found = []
		},
	},
}
</script>
