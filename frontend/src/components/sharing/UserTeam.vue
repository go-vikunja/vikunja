<template>
	<div>
		<p class="has-text-weight-bold">
			{{ $t('project.share.userTeam.shared', {type: shareTypeNames}) }}
		</p>
		<div v-if="userIsAdmin">
			<div class="field has-addons">
				<p
					class="control is-expanded"
					:class="{ 'is-loading': searchService.loading }"
				>
					<Multiselect
						v-model="sharable"
						:loading="searchService.loading"
						:placeholder="$t('misc.searchPlaceholder')"
						:search-results="found"
						:label="searchLabel"
						@search="find"
					>
						<template #searchResult="{option: result}">
							<User
								v-if="shareType === 'user'"
								:avatar-size="24"
								:show-username="true"
								:user="result"
							/>
							<span 
								v-else
								class="search-result"
							>
								{{ result.name }}
							</span>
						</template>
					</Multiselect>
				</p>
				<p class="control">
					<XButton @click="add()">
						{{ $t('project.share.share') }}
					</XButton>
				</p>
			</div>
		</div>

		<table
			v-if="sharables.length > 0"
			class="table has-actions is-striped is-hoverable is-fullwidth mb-4"
		>
			<tbody>
				<tr
					v-for="s in sharables"
					:key="s.id"
				>
					<template v-if="shareType === 'user'">
						<td>{{ getDisplayName(s) }}</td>
						<td>
							<template v-if="s.id === userInfo.id">
								<b class="is-success">{{ $t('project.share.userTeam.you') }}</b>
							</template>
						</td>
					</template>
					<template v-if="shareType === 'team'">
						<td>
							<RouterLink
								:to="{
									name: 'teams.edit',
									params: { id: s.id },
								}"
							>
								{{ s.name }}
							</RouterLink>
						</td>
					</template>
					<td class="type">
						<template v-if="s.right === RIGHTS.ADMIN">
							<span class="icon is-small">
								<Icon icon="lock" />
							</span>
							{{ $t('project.share.right.admin') }}
						</template>
						<template v-else-if="s.right === RIGHTS.READ_WRITE">
							<span class="icon is-small">
								<Icon icon="pen" />
							</span>
							{{ $t('project.share.right.readWrite') }}
						</template>
						<template v-else>
							<span class="icon is-small">
								<Icon icon="users" />
							</span>
							{{ $t('project.share.right.read') }}
						</template>
					</td>
					<td
						v-if="userIsAdmin"
						class="actions"
					>
						<div class="select">
							<select
								v-model="selectedRight[s.id]"
								class="mr-2"
								@change="toggleType(s)"
							>
								<option
									:selected="s.right === RIGHTS.READ"
									:value="RIGHTS.READ"
								>
									{{ $t('project.share.right.read') }}
								</option>
								<option
									:selected="s.right === RIGHTS.READ_WRITE"
									:value="RIGHTS.READ_WRITE"
								>
									{{ $t('project.share.right.readWrite') }}
								</option>
								<option
									:selected="s.right === RIGHTS.ADMIN"
									:value="RIGHTS.ADMIN"
								>
									{{ $t('project.share.right.admin') }}
								</option>
							</select>
						</div>
						<XButton
							class="is-danger"
							icon="trash-alt"
							@click="
								() => {
									sharable = s
									showDeleteModal = true
								}
							"
						/>
					</td>
				</tr>
			</tbody>
		</table>

		<Nothing v-else>
			{{ $t('project.share.userTeam.notShared', {type: shareTypeNames}) }}
		</Nothing>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteSharable()"
		>
			<template #header>
				<span>{{
					$t('project.share.userTeam.removeHeader', {type: shareTypeName, sharable: sharableName})
				}}</span>
			</template>
			<template #text>
				<p>{{ $t('project.share.userTeam.removeText', {type: shareTypeName, sharable: sharableName}) }}</p>
			</template>
		</Modal>
	</div>
</template>

<script lang="ts">
export default {name: 'UserTeamShare'}
</script>

<script setup lang="ts">
import {ref, reactive, computed, shallowReactive, type Ref} from 'vue'
import {useI18n} from 'vue-i18n'

import UserProjectService from '@/services/userProject'
import UserProjectModel from '@/models/userProject'
import type {IUserProject} from '@/modelTypes/IUserProject'

import UserService from '@/services/user'
import UserModel, { getDisplayName } from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

import TeamProjectService from '@/services/teamProject'
import TeamProjectModel from '@/models/teamProject'
import type { ITeamProject } from '@/modelTypes/ITeamProject'

import TeamService from '@/services/team'
import TeamModel from '@/models/team'
import type {ITeam} from '@/modelTypes/ITeam'


import {RIGHTS} from '@/constants/rights'
import Multiselect from '@/components/input/Multiselect.vue'
import Nothing from '@/components/misc/Nothing.vue'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import User from '@/components/misc/User.vue'

// FIXME: I think this whole thing can now only manage user/team sharing for projects? Maybe remove a little generalization?

const props = withDefaults(defineProps<{
	type?: 'project',
	shareType: 'user' | 'team',
	id: number,
	userIsAdmin?: boolean
}>(), {
	type: 'project',
	userIsAdmin: false,
})

const {t} = useI18n({useScope: 'global'})

// This user service is a userProjectService, depending on the type we are using
let stuffService: UserProjectService | TeamProjectService
let stuffModel: IUserProject | ITeamProject
let searchService: UserService | TeamService
let sharable: Ref<IUser | ITeam>

const searchLabel = ref('')
const selectedRight = ref({})


// This holds either teams or users who this namepace or project is shared with
const sharables = ref([])
const showDeleteModal = ref(false)

const authStore = useAuthStore()
const configStore = useConfigStore()
const userInfo = computed(() => authStore.info)

function createShareTypeNameComputed(count: number) {
	return computed(() => {
		if (props.shareType === 'user') {
			return t('project.share.userTeam.typeUser', count)
		}

		if (props.shareType === 'team') {
			return t('project.share.userTeam.typeTeam', count)
		}

		return ''
	})
}

const shareTypeNames = createShareTypeNameComputed(2)
const shareTypeName = createShareTypeNameComputed(1)

const sharableName = computed(() => {
	if (props.type === 'project') {
		return t('project.list.title')
	}

	return ''
})

if (props.shareType === 'user') {
	searchService = shallowReactive(new UserService())
	sharable = ref(new UserModel())
	searchLabel.value = 'username'

	if (props.type === 'project') {
		stuffService = shallowReactive(new UserProjectService())
		stuffModel = reactive(new UserProjectModel({projectId: props.id}))
	} else {
		throw new Error('Unknown type: ' + props.type)
	}
} else if (props.shareType === 'team') {
	searchService = new TeamService()
	sharable = ref(new TeamModel())
	searchLabel.value = 'name'

	if (props.type === 'project') {
		stuffService = shallowReactive(new TeamProjectService())
		stuffModel = reactive(new TeamProjectModel({projectId: props.id}))
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
		stuffModel.userId = sharable.value.username
	} else if (props.shareType === 'team') {
		stuffModel.teamId = sharable.value.id
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
	success({
		message: t('project.share.userTeam.removeSuccess', {
			type: shareTypeName.value,
			sharable: sharableName.value,
		}),
	})
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
		stuffModel.userId = sharable.value.username
	} else if (props.shareType === 'team') {
		stuffModel.teamId = sharable.value.id
	}

	await stuffService.create(stuffModel)
	success({message: t('project.share.userTeam.addedSuccess', {type: shareTypeName.value})})
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
	success({message: t('project.share.userTeam.updatedSuccess', {type: shareTypeName.value})})
}

const found = ref([])

const currentUserId = computed(() => authStore.info.id)
async function find(query: string) {
	if (query === '') {
		found.value = []
		return
	}

	// Include public teams here if we are sharing with teams and its enabled in the config
	let results = []
	if (props.shareType === 'team' && configStore.publicTeamsEnabled) {
		results = await searchService.getAll({}, {s: query, includePublic: true})
	} else {
		results = await searchService.getAll({}, {s: query})
	}

	found.value = results
		.filter(m => {
			if(props.shareType === 'user' && m.id === currentUserId.value) {
				return false
			}
			
			return typeof sharables.value.find(s => s.id === m.id) === 'undefined'
		})
}
</script>
