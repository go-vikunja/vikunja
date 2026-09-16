<template>
	<div>
		<h3 class="has-text-weight-bold share-heading">
			{{ $t('project.share.userTeam.shared', {type: shareTypeNames}) }}
		</h3>
		<div v-if="userIsAdmin">
			<div class="field has-addons">
				<p
					class="control is-expanded"
					:class="{ 'is-loading': searchLoading }"
				>
					<Multiselect
						v-model="sharable"
						:loading="searchLoading"
						:placeholder="$t('misc.searchPlaceholder')"
						:aria-label="$t('project.share.userTeam.search', {type: shareTypeName})"
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

		<div
			v-if="sharables.length > 0"
			class="has-horizontal-overflow mbe-4"
		>
			<table class="table has-actions is-striped is-hoverable is-fullwidth">
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
							<template v-if="s.permission === PERMISSIONS.ADMIN">
								<span class="icon is-small">
									<Icon icon="lock" />
								</span>
								{{ $t('project.share.permission.admin') }}
							</template>
							<template v-else-if="s.permission === PERMISSIONS.READ_WRITE">
								<span class="icon is-small">
									<Icon icon="pen" />
								</span>
								{{ $t('project.share.permission.readWrite') }}
							</template>
							<template v-else>
								<span class="icon is-small">
									<Icon icon="users" />
								</span>
								{{ $t('project.share.permission.read') }}
							</template>
						</td>
						<td
							v-if="userIsAdmin"
							class="actions"
						>
							<div class="select">
								<select
									v-model="selectedPermission[s.id]"
									class="mie-2"
									:aria-label="$t('project.share.userTeam.permissionFor', {sharable: shareType === 'user' ? getDisplayName(s) : s.name})"
									@change="toggleType(s)"
								>
									<option
										:selected="s.permission === PERMISSIONS.READ"
										:value="PERMISSIONS.READ"
									>
										{{ $t('project.share.permission.read') }}
									</option>
									<option
										:selected="s.permission === PERMISSIONS.READ_WRITE"
										:value="PERMISSIONS.READ_WRITE"
									>
										{{ $t('project.share.permission.readWrite') }}
									</option>
									<option
										:selected="s.permission === PERMISSIONS.ADMIN"
										:value="PERMISSIONS.ADMIN"
									>
										{{ $t('project.share.permission.admin') }}
									</option>
								</select>
							</div>
							<XButton
								danger
								icon="trash-alt"
								:aria-label="$t('project.share.userTeam.remove', {type: shareTypeName})"
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
		</div>

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


<script setup lang="ts">
import {ref, reactive, computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import UserProjectService from '@/services/userProject'
import UserProjectModel from '@/models/userProject'
import type {ProjectUser as IUserProject} from '@/client/generated'

import {useUserSearch} from '@/composables/useUserSearch'
import { getDisplayName } from '@/models/user'
import type {User as IUser} from '@/client/generated'

import TeamProjectService from '@/services/teamProject'
import TeamProjectModel from '@/models/teamProject'
import type {TeamProject as ITeamProject} from '@/client/generated'

import {useTeams} from '@/composables/useTeams'


import {PERMISSIONS} from '@/constants/permissions'
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

defineOptions({name: 'UserTeamShare'})

const {t} = useI18n({useScope: 'global'})

// This user service is a userProjectService, depending on the type we are using
let stuffService: UserProjectService | TeamProjectService
let stuffModel: IUserProject | ITeamProject
const configStore = useConfigStore()
const searchQuery = ref('')
const {teams: teamResults, isFetching: teamSearchLoading} = useTeams({search: searchQuery, includePublic: () => configStore.publicTeamsEnabled, enabled: () => props.shareType === 'team' && searchQuery.value !== ''})
const {users: userResults, isFetching: userSearchLoading} = useUserSearch(() => props.shareType === 'user' ? searchQuery.value : '')
const searchLoading = computed(() => teamSearchLoading.value || userSearchLoading.value)
const sharable = ref<IUser>({})

const searchLabel = ref('')
const selectedPermission = ref({})


// This holds either teams or users who this namepace or project is shared with
const sharables = ref([])
const showDeleteModal = ref(false)

const authStore = useAuthStore()
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
	sharable.value = {}
	searchLabel.value = 'username'

	if (props.type === 'project') {
		stuffService = shallowReactive(new UserProjectService())
		stuffModel = reactive(new UserProjectModel({projectId: props.id}))
	} else {
		throw new Error('Unknown type: ' + props.type)
	}
} else if (props.shareType === 'team') {
	sharable.value = {}
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
	sharables.value.forEach(({id, permission}) =>
		selectedPermission.value[id] = permission,
	)
}

async function deleteSharable() {
	if (props.shareType === 'user') {
		stuffModel.username = sharable.value.username
	} else if (props.shareType === 'team') {
		stuffModel.team_id = sharable.value.id
	}

	await stuffService.delete(stuffModel)
	showDeleteModal.value = false
	const idx = sharables.value.findIndex(s =>
		(props.shareType === 'user' && s.username === stuffModel.username) ||
		(props.shareType === 'team' && s.id === stuffModel.team_id),
	)
	if (idx !== -1) {
		sharables.value.splice(idx, 1)
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
	stuffModel.permission = PERMISSIONS.READ
	if (admin) {
		stuffModel.permission = PERMISSIONS.ADMIN
	}

	if (props.shareType === 'user') {
		stuffModel.username = sharable.value.username
	} else if (props.shareType === 'team') {
		stuffModel.team_id = sharable.value.id
	}

	await stuffService.create(stuffModel)
	success({message: t('project.share.userTeam.addedSuccess', {type: shareTypeName.value})})
	await load()
}

async function toggleType(sharable) {
	if (
		selectedPermission.value[sharable.id] !== PERMISSIONS.ADMIN &&
		selectedPermission.value[sharable.id] !== PERMISSIONS.READ &&
		selectedPermission.value[sharable.id] !== PERMISSIONS.READ_WRITE
	) {
		selectedPermission.value[sharable.id] = PERMISSIONS.READ
	}
	stuffModel.permission = selectedPermission.value[sharable.id]

	if (props.shareType === 'user') {
		stuffModel.username = sharable.username
	} else if (props.shareType === 'team') {
		stuffModel.team_id = sharable.id
	}

	const r = await stuffService.update(stuffModel)
	for (const sharableEntry of sharables.value) {
		if (
			(sharableEntry.username ===
				stuffModel.username &&
				props.shareType === 'user') ||
			(sharableEntry.id === stuffModel.team_id &&
				props.shareType === 'team')
		) {
			sharableEntry.permission = r.permission
		}
	}
	success({message: t('project.share.userTeam.updatedSuccess', {type: shareTypeName.value})})
}

const found = computed(() => (props.shareType === 'team' ? teamResults.value : userResults.value).filter(item => (props.shareType !== 'user' || item.id !== authStore.info?.id) && !sharables.value.some(shared => shared.id === item.id)))
function find(query: string) {
	searchQuery.value = query
}
</script>
