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
					<XButton
						:class="{'is-loading': isMutating}"
						:aria-disabled="!sharable || isMutating || undefined"
						@click="add()"
					>
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
								<template v-if="s.id === userInfo?.id">
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
									:aria-disabled="isMutating || undefined"
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
										shareToDelete = s
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
			:enabled="shareToDelete !== null"
			@close="shareToDelete = null"
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
import {ref, computed, watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useI18n} from 'vue-i18n'
import type {UserWithPermission, TeamWithPermission} from '@/client/generated'
import {projectUserSharesQuery, projectTeamSharesQuery, normalizeSharePermission, useCreateProjectUserShareMutation, useUpdateProjectUserShareMutation, useDeleteProjectUserShareMutation, useCreateProjectTeamShareMutation, useUpdateProjectTeamShareMutation, useDeleteProjectTeamShareMutation} from '@/client/queries/projectShares'
import {useTeams} from '@/composables/useTeams'
import {useUserSearch} from '@/composables/useUserSearch'
import {getDisplayName} from '@/models/user'
import {PERMISSIONS, type Permission} from '@/constants/permissions'
import Multiselect from '@/components/input/Multiselect.vue'
import Nothing from '@/components/misc/Nothing.vue'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import User from '@/components/misc/User.vue'

const props = withDefaults(defineProps<{
	shareType: 'user' | 'team',
	id: number,
	userIsAdmin?: boolean,
}>(), {userIsAdmin: false})
defineOptions({name: 'UserTeamShare'})

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const configStore = useConfigStore()
const userInfo = computed(() => authStore.info)
type Share = UserWithPermission | TeamWithPermission
type IdentifiedShare = Share & Required<Pick<Share, 'id'>>
const usersQuery = useQuery(computed(() => ({...projectUserSharesQuery(props.id), enabled: props.id > 0 && props.shareType === 'user'})))
const teamsQuery = useQuery(computed(() => ({...projectTeamSharesQuery(props.id), enabled: props.id > 0 && props.shareType === 'team'})))
const sharables = computed(() => (props.shareType === 'user' ? usersQuery.data.value ?? [] : teamsQuery.data.value ?? []).filter((item): item is IdentifiedShare => typeof item.id === 'number'))
const searchQuery = ref('')
const {teams: teamResults, isFetching: teamSearchLoading} = useTeams({search: searchQuery, includePublic: () => configStore.publicTeamsEnabled, enabled: () => props.shareType === 'team' && searchQuery.value !== ''})
const {users: userResults, isFetching: userSearchLoading} = useUserSearch(() => props.shareType === 'user' ? searchQuery.value : '')
const searchLoading = computed(() => teamSearchLoading.value || userSearchLoading.value)
const found = computed(() => (props.shareType === 'team' ? teamResults.value : userResults.value).filter(item => (props.shareType !== 'user' || item.id !== userInfo.value?.id) && !sharables.value.some(shared => shared.id === item.id)))
const searchLabel = computed(() => props.shareType === 'user' ? 'username' : 'name')
const shareTypeName = computed(() => t(props.shareType === 'user' ? 'project.share.userTeam.typeUser' : 'project.share.userTeam.typeTeam', 1))
const shareTypeNames = computed(() => t(props.shareType === 'user' ? 'project.share.userTeam.typeUser' : 'project.share.userTeam.typeTeam', 2))
const sharableName = computed(() => t('project.list.title'))
const sharable = ref<Share | null>(null)
const selectedPermission = ref<Record<number, Permission>>({})
const shareToDelete = ref<IdentifiedShare | null>(null)
watch(sharables, shares => {
	selectedPermission.value = Object.fromEntries(shares.map(share => [share.id, normalizeSharePermission(share.permission)]))
}, {immediate: true})
watch(() => [props.id, props.shareType], () => {
	sharable.value = null
	searchQuery.value = ''
	shareToDelete.value = null
})

const createUser = useCreateProjectUserShareMutation()
const updateUser = useUpdateProjectUserShareMutation()
const deleteUser = useDeleteProjectUserShareMutation()
const createTeam = useCreateProjectTeamShareMutation()
const updateTeam = useUpdateProjectTeamShareMutation()
const deleteTeam = useDeleteProjectTeamShareMutation()
const isMutating = computed(() => createUser.isPending.value || updateUser.isPending.value || deleteUser.isPending.value || createTeam.isPending.value || updateTeam.isPending.value || deleteTeam.isPending.value)

async function add() {
	const selected = sharable.value
	const projectId = props.id
	if (!selected || isMutating.value) return
	try {
		if (props.shareType === 'user' && 'username' in selected && selected.username) {
			await createUser.mutateAsync({projectId, username: selected.username})
		} else if (props.shareType === 'team' && selected.id !== undefined) {
			await createTeam.mutateAsync({projectId, teamId: selected.id})
		} else return
	} catch {
		return
	}
	if (props.id === projectId && sharable.value === selected) {
		sharable.value = null
		searchQuery.value = ''
	}
}

function resetSelectedPermission(shareId: number) {
	selectedPermission.value[shareId] = normalizeSharePermission(sharables.value.find(item => item.id === shareId)?.permission)
}

function toggleType(share: IdentifiedShare) {
	const projectId = props.id
	const reset = () => {
		if (props.id === projectId) resetSelectedPermission(share.id)
	}
	if (isMutating.value) {
		reset()
		return
	}
	const permission = normalizeSharePermission(selectedPermission.value[share.id])
	if (props.shareType === 'user' && 'username' in share && share.username) {
		updateUser.mutate({projectId, username: share.username, permission}, {onSettled: reset})
	} else if (props.shareType === 'team') {
		updateTeam.mutate({projectId, teamId: share.id, permission}, {onSettled: reset})
	} else {
		reset()
	}
}

function deleteSharable() {
	const selected = shareToDelete.value
	const projectId = props.id
	if (!selected || isMutating.value) return
	const close = () => {
		if (shareToDelete.value === selected) shareToDelete.value = null
	}
	if (props.shareType === 'user' && 'username' in selected && selected.username) {
		deleteUser.mutate({projectId, username: selected.username}, {onSettled: close})
	} else if (props.shareType === 'team') {
		deleteTeam.mutate({projectId, teamId: selected.id}, {onSettled: close})
	}
}

function find(query: string) {
	searchQuery.value = query
}
</script>
