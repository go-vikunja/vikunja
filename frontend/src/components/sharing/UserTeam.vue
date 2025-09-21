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
								:user="result as IUser"
							/>
							<span 
								v-else
								class="search-result"
							>
								{{ (result as ITeam).name }}
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
			class="table has-actions is-striped is-hoverable is-fullwidth mbe-4"
		>
			<tbody>
				<tr
					v-for="s in sharables"
					:key="s.id"
				>
					<template v-if="shareType === 'user'">
						<td>{{ getDisplayName(s as unknown as IUser) }}</td>
						<td>
							<template v-if="userInfo && s.id === userInfo.id">
								<b class="is-success">{{ $t('project.share.userTeam.you') }}</b>
							</template>
						</td>
					</template>
					<template v-if="shareType === 'team'">
						<td>
							<RouterLink
								:to="{
									name: 'teams.edit',
									params: { id: (s as ITeamProject).id as string | number },
								}"
							>
								{{ (s as ITeamProject).name }}
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
							class="is-danger"
							icon="trash-alt"
							@click="
								() => {
									sharable = s as unknown as IUser | ITeam
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


import {PERMISSIONS, type Permission} from '@/constants/permissions'
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
let searchService: UserService | TeamService
let sharable: Ref<IUser | ITeam>

const searchLabel = ref('')
const selectedPermission = ref<Record<number, Permission>>({})


type SharableItem = (IUserProject | ITeamProject) & {id: number}

// This holds either teams or users who this namepace or project is shared with
const sharables = ref<SharableItem[]>([])
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
	sharable = ref<IUser | ITeam>(new UserModel())
	searchLabel.value = 'username'

	if (props.type === 'project') {
		stuffService = shallowReactive(new UserProjectService())
		stuffModel = reactive(new UserProjectModel({projectId: props.id}))
	} else {
		throw new Error('Unknown type: ' + props.type)
	}
} else if (props.shareType === 'team') {
	searchService = new TeamService()
	sharable = ref<IUser | ITeam>(new TeamModel())
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
	if (props.shareType === 'user') {
		sharables.value = await (stuffService as UserProjectService).getAll(stuffModel as IUserProject) as SharableItem[]
	} else if (props.shareType === 'team') {
		sharables.value = await (stuffService as TeamProjectService).getAll(stuffModel as ITeamProject) as SharableItem[]
	}
	sharables.value.forEach((sharable) =>
		selectedPermission.value[sharable.id] = sharable.permission,
	)
}

async function deleteSharable() {
	if (props.shareType === 'user') {
		(stuffModel as IUserProject).username = (sharable.value as IUser).username || ''
	} else if (props.shareType === 'team') {
		(stuffModel as ITeamProject).teamId = (sharable.value as ITeam).id || 0
	}

	if (props.shareType === 'user') {
		await (stuffService as UserProjectService).delete(stuffModel as IUserProject)
	} else if (props.shareType === 'team') {
		await (stuffService as TeamProjectService).delete(stuffModel as ITeamProject)
	}
	showDeleteModal.value = false
	for (const i in sharables.value) {
		if (
			((sharables.value[i] as IUserProject).username === (stuffModel as IUserProject).username && props.shareType === 'user') ||
			((sharables.value[i] as ITeamProject).teamId === (stuffModel as ITeamProject).teamId && props.shareType === 'team')
		) {
			sharables.value.splice(parseInt(i), 1)
		}
	}
	success({
		message: t('project.share.userTeam.removeSuccess', {
			type: shareTypeName.value,
			sharable: sharableName.value,
		}),
	})
}

async function add(admin?: boolean) {
	if (admin === null || admin === undefined) {
		admin = false
	}
	stuffModel.permission = PERMISSIONS.READ
	if (admin) {
		stuffModel.permission = PERMISSIONS.ADMIN
	}

	if (props.shareType === 'user') {
		(stuffModel as IUserProject).username = (sharable.value as IUser).username || ''
	} else if (props.shareType === 'team') {
		(stuffModel as ITeamProject).teamId = (sharable.value as ITeam).id || 0
	}

	if (props.shareType === 'user') {
		await (stuffService as UserProjectService).create(stuffModel as IUserProject)
	} else if (props.shareType === 'team') {
		await (stuffService as TeamProjectService).create(stuffModel as ITeamProject)
	}
	success({message: t('project.share.userTeam.addedSuccess', {type: shareTypeName.value})})
	await load()
}

async function toggleType(sharable: SharableItem) {
	const sharableId = sharable.id
	if (
		selectedPermission.value[sharableId] !== PERMISSIONS.ADMIN &&
		selectedPermission.value[sharableId] !== PERMISSIONS.READ &&
		selectedPermission.value[sharableId] !== PERMISSIONS.READ_WRITE
	) {
		selectedPermission.value[sharableId] = PERMISSIONS.READ
	}
	stuffModel.permission = selectedPermission.value[sharableId]

	if (props.shareType === 'user') {
		(stuffModel as IUserProject).username = (sharable as unknown as IUser).username || ''
	} else if (props.shareType === 'team') {
		(stuffModel as ITeamProject).teamId = (sharable as unknown as ITeam).id || 0
	}

	let r
	if (props.shareType === 'user') {
		r = await (stuffService as UserProjectService).update(stuffModel as IUserProject)
	} else if (props.shareType === 'team') {
		r = await (stuffService as TeamProjectService).update(stuffModel as ITeamProject)
	}
	for (const i in sharables.value) {
		if (
			((sharables.value[i] as IUserProject).username ===
				(stuffModel as IUserProject).username &&
				props.shareType === 'user') ||
			((sharables.value[i] as ITeamProject).teamId === (stuffModel as ITeamProject).teamId &&
				props.shareType === 'team')
		) {
			if (r && sharables.value[i]) {
				sharables.value[i].permission = r.permission
			}
		}
	}
	success({message: t('project.share.userTeam.updatedSuccess', {type: shareTypeName.value})})
}

const found = ref<(IUser | ITeam)[]>([])

const currentUserId = computed(() => authStore.info?.id)
async function find(query: string) {
	if (query === '') {
		found.value = []
		return
	}

	// Include public teams here if we are sharing with teams and its enabled in the config
	let results = []
	if (props.shareType === 'team' && configStore.publicTeamsEnabled) {
		results = await (searchService as TeamService).getAll(undefined, {s: query, includePublic: true})
	} else if (props.shareType === 'team') {
		results = await (searchService as TeamService).getAll(undefined, {s: query})
	} else {
		results = await (searchService as UserService).getAll(undefined, {s: query})
	}

	found.value = results
		.filter(m => {
			if(props.shareType === 'user' && m.id === currentUserId.value) {
				return false
			}

			return typeof sharables.value.find((s) => s.id === m.id) === 'undefined'
		})
}
</script>
