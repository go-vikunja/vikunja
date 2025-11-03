<template>
	<div
		class="loader-container is-max-width-desktop"
		:class="{ 'is-loading': teamService.loading }"
	>
		<Card
			v-if="userIsAdmin && !team.oidcId"
			class="is-fullwidth"
			:title="title"
		>
			<form @submit.prevent="save()">
				<div class="field">
					<label
						class="label"
						for="teamtext"
					>{{ $t('team.attributes.name') }}</label>
					<div class="control">
						<input
							id="teamtext"
							v-model="team.name"
							v-focus
							:class="{ disabled: teamMemberService.loading }"
							:disabled="teamMemberService.loading || undefined"
							class="input"
							:placeholder="$t('team.attributes.namePlaceholder')"
							type="text"
						>
					</div>
				</div>
				<p
					v-if="showErrorTeamnameRequired && team.name === ''"
					class="help is-danger"
				>
					{{ $t('team.attributes.nameRequired') }}
				</p>
				<div
					v-if="configStore.publicTeamsEnabled"
					class="field"
				>
					<label
						class="label"
						for="teamIsPublic"
					>{{ $t('team.attributes.isPublic') }}</label>
					<div
						class="control is-expanded"
						:class="{ 'is-loading': teamService.loading }"
					>
						<FancyCheckbox
							v-model="team.isPublic"
							:disabled="teamMemberService.loading || undefined"
							:class="{ 'disabled': teamService.loading }"
						>
							{{ $t('team.attributes.isPublicDescription') }}
						</FancyCheckbox>
					</div>
				</div>
				<div class="field">
					<label
						class="label"
						for="teamdescription"
					>{{ $t('team.attributes.description') }}</label>
					<div class="control">
						<Editor
							id="teamdescription"
							v-model="team.description"
							:class="{ disabled: teamService.loading }"
							:disabled="teamService.loading"
							:placeholder="$t('team.attributes.descriptionPlaceholder')"
						/>
					</div>
				</div>
			</form>

			<div class="field has-addons mbs-4">
				<div class="control is-fullwidth">
					<XButton
						:loading="teamService.loading"
						class="is-fullwidth"
						@click="save()"
					>
						{{ $t('misc.save') }}
					</XButton>
				</div>
				<div class="control">
					<XButton
						:loading="teamService.loading"
						class="is-danger"
						icon="trash-alt"
						@click="showDeleteModal = true"
					/>
				</div>
			</div>
		</Card>

		<Card
			class="is-fullwidth has-overflow"
			:title="$t('team.edit.members')"
			:padding="false"
		>
			<form
				v-if="userIsAdmin && !team.oidcId"
				class="p-4"
				@submit.prevent="addUser"
			>
				<div class="field has-addons">
					<div class="control is-expanded">
						<Multiselect
							v-model="newMember"
							:loading="userService.loading"
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
			<table class="table has-actions is-striped is-hoverable is-fullwidth">
				<tbody>
					<tr
						v-for="m in team?.members"
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
							<template v-if="m.id === userInfo.id">
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
								v-if="m.id !== userInfo.id"
								:loading="teamMemberService.loading"
								class="mie-2"
								@click="() => toggleUserType(m)"
							>
								{{ m.admin ? $t('team.edit.makeMember') : $t('team.edit.makeAdmin') }}
							</XButton>
							<XButton
								v-if="m.id !== userInfo.id"
								:loading="teamMemberService.loading"
								class="is-danger"
								icon="trash-alt"
								@click="() => {memberToDelete = m; showUserDeleteModal = true}"
							/>
						</td>
					</tr>
				</tbody>
			</table>
		</Card>

		<XButton
			v-if="team && !team.externalId"
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

import Editor from '@/components/input/AsyncEditor'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'

import TeamService from '@/services/team'
import TeamMemberService from '@/services/teamMember'
import UserService from '@/services/user'

import {PERMISSIONS as Permissions} from '@/constants/permissions'

import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'

import type {ITeam} from '@/modelTypes/ITeam'
import type {IUser} from '@/modelTypes/IUser'
import type {ITeamMember} from '@/modelTypes/ITeamMember'

const authStore = useAuthStore()
const configStore = useConfigStore()
const route = useRoute()
const router = useRouter()
const {t} = useI18n({useScope: 'global'})

const userIsAdmin = computed(() => {
	return (
		team.value &&
		team.value.maxPermission &&
		team.value.maxPermission > Permissions.READ
	)
})
const userInfo = computed(() => authStore.info)

const teamService = ref<TeamService>(new TeamService())
const teamMemberService = ref<TeamMemberService>(new TeamMemberService())
const userService = ref<UserService>(new UserService())

const team = ref<ITeam>()
const teamId = computed(() => Number(route.params.id))
const memberToDelete = ref<ITeamMember>()
const newMember = ref<IUser>()
const foundUsers = ref<IUser[]>()

const showDeleteModal = ref(false)
const showUserDeleteModal = ref(false)
const showLeaveModal = ref(false)
const showErrorTeamnameRequired = ref(false)
const showMustSelectUserError = ref(false)

const title = ref('')

loadTeam()

async function loadTeam() {
	team.value = await teamService.value.get({id: teamId.value})
	title.value = t('team.edit.title', {team: team.value?.name})
	useTitle(() => title.value)
}

async function save() {
	if (team.value?.name === '') {
		showErrorTeamnameRequired.value = true
		return
	}
	showErrorTeamnameRequired.value = false

	team.value = await teamService.value.update(team.value)
	success({message: t('team.edit.success')})
}

async function deleteTeam() {
	await teamService.value.delete(team.value)
	success({message: t('team.edit.delete.success')})
	router.push({name: 'teams.index'})
}

async function deleteMember() {
	try {
		await teamMemberService.value.delete({
			teamId: teamId.value,
			username: memberToDelete.value.username,
		})
		success({message: t('team.edit.deleteUser.success')})
		await loadTeam()
	} finally {
		showUserDeleteModal.value = false
	}
}

async function addUser() {
	showMustSelectUserError.value = false
	if(!newMember.value) {
		showMustSelectUserError.value = true
		return
	}
	await teamMemberService.value.create({
		teamId: teamId.value,
		username: newMember.value.username,
	})
	newMember.value = null
	await loadTeam()
	success({message: t('team.edit.userAddedSuccess')})
}

async function toggleUserType(member: ITeamMember) {
	// FIXME: direct manipulation
	member.admin = !member.admin
	member.teamId = teamId.value
	const r = await teamMemberService.value.update(member)
	for (const tm in team.value.members) {
		if (team.value.members[tm].id === member.id) {
			team.value.members[tm].admin = r.admin
			break
		}
	}
	success({
		message: member.admin ?
			t('team.edit.madeAdmin') :
			t('team.edit.madeMember'),
	})
}

async function findUser(query: string) {
	if (query === '') {
		foundUsers.value = []
		return
	}

	const users = await userService.value.getAll({}, {s: query})
	foundUsers.value = users.filter((u: IUser) => u.id !== userInfo.value.id)
}

async function leave() {
	try {
		await teamMemberService.value.delete({
			teamId: teamId.value,
			username: userInfo.value.username,
		})
		success({message: t('team.edit.leave.success')})
		await router.push({name: 'home'})
	} finally {
		showUserDeleteModal.value = false
	}
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
