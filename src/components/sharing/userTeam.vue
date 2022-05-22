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
					<Multiselect
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
							class="mr-2"
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

<script lang="ts">
import {defineComponent, ShallowReactive, shallowReactive} from 'vue'
export default defineComponent({ name: 'userTeamShare' })
</script>

<script setup lang="ts">
import {ref, reactive, computed} from 'vue'
import type {PropType} from 'vue'
import {useStore} from 'vuex'
import {useI18n} from 'vue-i18n'

import UserNamespaceService from '@/services/userNamespace'
import UserNamespaceModel from '@/models/userNamespace'
import UserListModel from '@/models/userList'
import UserListService from '@/services/userList'
import UserService from '@/services/user'
import UserModel from '@/models/user'

import TeamNamespaceService from '@/services/teamNamespace'
import TeamNamespaceModel from '@/models/teamNamespace'
import TeamListModel from '@/models/teamList'
import TeamListService from '@/services/teamList'
import TeamService from '@/services/team'
import TeamModel from '@/models/team'

import RIGHTS from '@/models/constants/rights.json'
import Multiselect from '@/components/input/multiselect.vue'
import Nothing from '@/components/misc/nothing.vue'
import { success } from '@/message'

const props = defineProps({
	type: {
		type: String as PropType<'list' | 'namespace'>,
		default: '',
	},
	shareType: {
		type: String as PropType<'user' | 'team' | 'namespace'>,
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
})

const {t} = useI18n()

// This user service is either a userNamespaceService or a userListService, depending on the type we are using
let stuffService: ShallowReactive<UserNamespaceService | UserListService | TeamListService | TeamNamespaceService>
let stuffModel: UserNamespaceModel | UserListModel | TeamListModel | TeamNamespaceModel
let searchService: ShallowReactive<UserService | TeamService>
let sharable: UserModel | TeamModel
 
const searchLabel = ref('')
const selectedRight = ref({})


// This holds either teams or users who this namepace or list is shared with
const sharables = ref([])
const showDeleteModal = ref(false)


const store = useStore()
const userInfo = computed(() => store.state.auth.info)

function createShareTypeNameComputed(count: number) {
	return computed(() => {
		if (props.shareType === 'user') {
			return t('list.share.userTeam.typeUser', count)
		}

		if (props.shareType === 'team') {
			return t('list.share.userTeam.typeTeam', count)
		}

		return ''
	})
}

const shareTypeNames = createShareTypeNameComputed(2)
const shareTypeName = createShareTypeNameComputed(1)

const sharableName = computed(() => {
	if (props.type === 'list') {
		return t('list.list.title')
	}

	if (props.shareType === 'namespace') {
		return t('namespace.namespace')
	}

	return ''
})

if (props.shareType === 'user') {
	searchService = shallowReactive(new UserService())
	sharable = reactive(new UserModel())
	searchLabel.value = 'username'

	if (props.type === 'list') {
		stuffService = shallowReactive(new UserListService())
		stuffModel = reactive(new UserListModel({listId: props.id}))
	} else if (props.type === 'namespace') {
		stuffService = shallowReactive(new UserNamespaceService())
		stuffModel = reactive(new UserNamespaceModel({
			namespaceId: props.id,
		}))
	} else {
		throw new Error('Unknown type: ' + props.type)
	}
} else if (props.shareType === 'team') {
	searchService = new TeamService()
	sharable = reactive(new TeamModel())
	searchLabel.value = 'name'

	if (props.type === 'list') {
		stuffService = shallowReactive(new TeamListService())
		stuffModel = reactive(new TeamListModel({listId: props.id}))
	} else if (props.type === 'namespace') {
		stuffService = shallowReactive(new TeamNamespaceService())
		stuffModel = reactive(new TeamNamespaceModel({
			namespaceId: props.id,
		}))
	} else {
		throw new Error('Unknown type: ' + props.type)
	}
} else {
	throw new Error('Unkown share type')
}

load()

async function load() {
	sharables.value = await stuffService.getAll(stuffModel)
	sharables.value.forEach(({id, right}) =>
		selectedRight.value[id] = right,
	)
}

async function deleteSharable() {
	if (props.shareType === 'user') {
		stuffModel.userId = sharable.username
	} else if (props.shareType === 'team') {
		stuffModel.teamId = sharable.id
	}

	await stuffService.delete(stuffModel)
	showDeleteModal.value = false
	for (const i in sharables.value) {
		if (
			(sharables.value[i].username === stuffModel.userId && props.shareType === 'user') ||
			(sharables.value[i].id === stuffModel.teamId && props.shareType === 'team')
		) {
			sharables.value.splice(i, 1)
		}
	}
	success({message: t('list.share.userTeam.removeSuccess', {
		type: shareTypeName.value,
		sharable: sharableName.value,
	})})
}

async function add(admin) {
	if (admin === null) {
		admin = false
	}
	stuffModel.right = RIGHTS.READ
	if (admin) {
		stuffModel.right = RIGHTS.ADMIN
	}

	if (props.shareType === 'user') {
		stuffModel.userId = sharable.username
	} else if (props.shareType === 'team') {
		stuffModel.teamId = sharable.id
	}

	await stuffService.create(stuffModel)
	success({message: t('list.share.userTeam.addedSuccess', {type: shareTypeName.value})})
	await load()
}

async function toggleType(sharable) {
	if (
		selectedRight.value[sharable.id] !== RIGHTS.ADMIN &&
		selectedRight.value[sharable.id] !== RIGHTS.READ &&
		selectedRight.value[sharable.id] !== RIGHTS.READ_WRITE
	) {
		selectedRight.value[sharable.id] = RIGHTS.READ
	}
	stuffModel.right = selectedRight.value[sharable.id]

	if (props.shareType === 'user') {
		stuffModel.userId = sharable.username
	} else if (props.shareType === 'team') {
		stuffModel.teamId = sharable.id
	}

	const r = await stuffService.update(stuffModel)
	for (const i in sharables.value) {
		if (
			(sharables.value[i].username ===
				stuffModel.userId &&
				props.shareType === 'user') ||
			(sharables.value[i].id === stuffModel.teamId &&
				props.shareType === 'team')
		) {
			sharables.value[i].right = r.right
		}
	}
	success({message: t('list.share.userTeam.updatedSuccess', {type: shareTypeName.value})})
}

const found = ref([])
async function find(query) {
	if (query === '') {
		clearAll()
		return
	}
	found.value = await searchService.getAll({}, {s: query})
}

function clearAll() {
	found.value = []
}
</script>

<style lang="scss" scoped>
@include modal-transition();
</style>