<template>
	<div
		class="loader-container is-max-width-desktop"
		:class="{ 'is-loading': teamLoading }"
	>
		<EditTeamForm
			v-if="team && userIsAdmin && !team.external_id"
			:key="team.id"
			:team="team"
			:loading="teamLoading"
			:members-loading="membersLoading"
			@save="save"
			@delete="showDeleteModal = true"
		/>

		<Card
			class="is-fullwidth has-overflow"
			:title="$t('team.edit.members')"
			:padding="false"
		>
			<form
				v-if="userIsAdmin && !team?.external_id"
				class="p-4"
				@submit.prevent="addUser"
			>
				<div class="field has-addons">
					<div class="control is-expanded">
						<Multiselect
							v-model="newMember"
							:loading="usersLoading"
							:placeholder="$t('team.edit.search')"
							:search-results="foundUsers"
							label="username"
							@search="findUser"
						>
							<template #searchResult="{option: user}">
								<User
									:avatar-size="24"
									:user="user"
									class="m-0"
								/>
							</template>
						</Multiselect>
					</div>
					<div class="control">
						<XButton
							icon="plus"
							@click="addUser"
						>
							{{ $t('team.edit.addUser') }}
						</XButton>
					</div>
				</div>
				<p
					v-if="showMustSelectUserError"
					class="help is-danger"
				>
					{{ $t('team.edit.mustSelectUser') }}
				</p>
			</form>
			<div class="has-horizontal-overflow">
				<table class="table has-actions is-striped is-hoverable is-fullwidth">
					<tbody>
						<tr
							v-for="m in sortedMembers"
							:key="m.id"
						>
							<td>
								<User
									:avatar-size="24"
									:user="m"
									class="m-0"
								/>
							</td>
							<td>
								<template v-if="m.id === userInfo?.id">
									<b class="is-success">You</b>
								</template>
							</td>
							<td class="type">
								<template v-if="m.admin">
									<span class="icon is-small">
										<Icon icon="lock" />
									</span>
									{{ $t('team.attributes.admin') }}
								</template>
								<template v-else>
									<span class="icon is-small">
										<Icon icon="user" />
									</span>
									{{ $t('team.attributes.member') }}
								</template>
							</td>
							<td
								v-if="userIsAdmin"
								class="actions"
							>
								<XButton
									v-if="m.id !== userInfo?.id"
									:loading="membersLoading"
									class="mie-2"
									@click="() => toggleUserType(m)"
								>
									{{ m.admin ? $t('team.edit.makeMember') : $t('team.edit.makeAdmin') }}
								</XButton>
								<XButton
									v-if="m.id !== userInfo?.id"
									:loading="membersLoading"
									danger
									icon="trash-alt"
									:aria-label="$t('team.edit.deleteUser.header')"
									@click="() => {memberToDelete = m; showUserDeleteModal = true}"
								/>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</Card>

		<XButton
			v-if="team && !team.external_id"
			class="is-fullwidth is-danger"
			@click="showLeaveModal = true"
		>
			{{ $t('team.edit.leave.title') }}
		</XButton>

		<!-- Leave team modal -->
		<Modal
			v-if="showLeaveModal"
			@close="showLeaveModal = false"
			@submit="leave()"
		>
			<template #header>
				<span>{{ $t('team.edit.leave.title') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('team.edit.leave.text1') }}<br>
					{{ $t('team.edit.leave.text2') }}
				</p>
			</template>
		</Modal>

		<!-- Team delete modal -->
		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteTeam()"
		>
			<template #header>
				<span>{{ $t('team.edit.delete.header') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('team.edit.delete.text1') }}<br>
					{{ $t('team.edit.delete.text2') }}
				</p>
			</template>
		</Modal>

		<!-- User delete modal -->
		<Modal
			:enabled="showUserDeleteModal"
			@close="showUserDeleteModal = false"
			@submit="deleteMember()"
		>
			<template #header>
				<span>{{ $t('team.edit.deleteUser.header') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('team.edit.deleteUser.text1') }}<br>
					{{ $t('team.edit.deleteUser.text2') }}
				</p>
			</template>
		</Modal>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'
import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'
import EditTeamForm from '@/components/teams/EditTeamForm.vue'
import {getDisplayName} from '@/models/user'
import {useUpdateTeamMutation, useDeleteTeamMutation, useAddTeamMemberMutation, useRemoveTeamMemberMutation, useLeaveTeamMutation, useToggleTeamMemberAdminMutation} from '@/client/queries/teams'
import {useTeam} from '@/composables/useTeams'
import {useUserSearch} from '@/composables/useUserSearch'
import {PERMISSIONS} from '@/constants/permissions'
import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'
import type {User as ApiUser, TeamUser, TeamWritable} from '@/client/generated'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const {t} = useI18n({useScope: 'global'})
const teamId = computed(() => Number(route.params.id))
const {data: team, isFetching} = useTeam(teamId)

const userIsAdmin = computed(() => (team.value?.max_permission ?? PERMISSIONS.READ) === PERMISSIONS.ADMIN)
const userInfo = computed(() => authStore.info)
const sortedMembers = computed(() => [...(team.value?.members ?? [])].sort((a, b) => getDisplayName(a).localeCompare(getDisplayName(b), undefined, {sensitivity: 'base'})))
const updateMutation = useUpdateTeamMutation()
const deleteMutation = useDeleteTeamMutation()
const addMemberMutation = useAddTeamMemberMutation()
const removeMemberMutation = useRemoveTeamMemberMutation()
const leaveMutation = useLeaveTeamMutation()
const toggleAdminMutation = useToggleTeamMemberAdminMutation()
const teamLoading = computed(() => isFetching.value || updateMutation.isPending.value || deleteMutation.isPending.value)
const membersLoading = computed(() => addMemberMutation.isPending.value || removeMemberMutation.isPending.value || toggleAdminMutation.isPending.value)
const userSearch = ref('')
const {users: userResults, isFetching: usersLoading} = useUserSearch(userSearch)
const foundUsers = computed(() => userResults.value.filter(u => u.id !== userInfo.value?.id && !team.value?.members?.some(member => member.id === u.id)))
const memberToDelete = ref<TeamUser>()
const newMember = ref<ApiUser | null>(null)
const showDeleteModal = ref(false)
const showUserDeleteModal = ref(false)
const showLeaveModal = ref(false)
const showMustSelectUserError = ref(false)
const title = computed(() => t('team.edit.title', {team: team.value?.name ?? ''}))
useTitle(title)

function save(draft: Required<TeamWritable>) {
	updateMutation.mutate({id: teamId.value, team: draft})
}

async function deleteTeam() {
	try {
		await deleteMutation.mutateAsync(teamId.value)
	} catch {
		return
	} finally {
		showDeleteModal.value = false
	}
	await router.push({name: 'teams.index'})
}

function deleteMember() {
	if (!memberToDelete.value?.username) return
	removeMemberMutation.mutate({teamId: teamId.value, username: memberToDelete.value.username}, {
		onSettled: () => {
			showUserDeleteModal.value = false
		},
	})
}

async function addUser() {
	showMustSelectUserError.value = !newMember.value?.username
	if (!newMember.value?.username) return
	try {
		await addMemberMutation.mutateAsync({teamId: teamId.value, username: newMember.value.username})
	} catch {
		return
	}
	newMember.value = null
	userSearch.value = ''
}

function toggleUserType(member: TeamUser) {
	if (!member.username) return
	toggleAdminMutation.mutate({teamId: teamId.value, username: member.username})
}

function findUser(query: string) {
	userSearch.value = query
}

async function leave() {
	if (!userInfo.value?.username) return
	try {
		await leaveMutation.mutateAsync({teamId: teamId.value, username: userInfo.value.username})
	} catch {
		return
	} finally {
		showLeaveModal.value = false
	}
	await router.push({name: 'home'})
}
</script>

<style lang="scss" scoped>
.card.is-fullwidth {
	margin-block-end: 1rem;

	.content {
		padding: 0;
	}
}
</style>
