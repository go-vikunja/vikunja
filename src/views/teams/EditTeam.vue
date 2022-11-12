<template>
	<div
		class="loader-container is-max-width-desktop"
		:class="{ 'is-loading': teamService.loading }"
	>
		<card class="is-fullwidth" v-if="userIsAdmin" :title="title">
			<form @submit.prevent="save()">
				<div class="field">
					<label class="label" for="teamtext">{{ $t('team.attributes.name') }}</label>
					<div class="control">
						<input
							:class="{ disabled: teamMemberService.loading }"
							:disabled="teamMemberService.loading || undefined"
							class="input"
							id="teamtext"
							:placeholder="$t('team.attributes.namePlaceholder')"
							type="text"
							v-focus
							v-model="team.name"
						/>
					</div>
				</div>
				<p
					class="help is-danger"
					v-if="showErrorTeamnameRequired && team.name === ''"
				>
					{{ $t('team.attributes.nameRequired') }}
				</p>
				<div class="field">
					<label class="label" for="teamdescription">{{ $t('team.attributes.description') }}</label>
					<div class="control">
						<editor
							:class="{ disabled: teamService.loading }"
							:disabled="teamService.loading"
							:preview-is-default="false"
							id="teamdescription"
							:placeholder="$t('team.attributes.descriptionPlaceholder')"
							v-model="team.description"
						/>
					</div>
				</div>
			</form>

			<div class="field has-addons mt-4">
				<div class="control is-fullwidth">
					<x-button
						@click="save()"
						:loading="teamService.loading"
						class="is-fullwidth"
					>
						{{ $t('misc.save') }}
					</x-button>
				</div>
				<div class="control">
					<x-button
						@click="showDeleteModal = true"
						:loading="teamService.loading"
						class="is-danger"
						icon="trash-alt"
					/>
				</div>
			</div>
		</card>

		<card class="is-fullwidth has-overflow" :title="$t('team.edit.members')" :padding="false">
			<div class="p-4" v-if="userIsAdmin">
				<div class="field has-addons">
					<div class="control is-expanded">
						<multiselect
							:loading="userService.loading"
							:placeholder="$t('team.edit.search')"
							@search="findUser"
							:search-results="foundUsers"
							label="username"
							v-model="newMember"
						/>
					</div>
					<div class="control">
						<x-button @click="addUser" icon="plus">
							{{ $t('team.edit.addUser') }}
						</x-button>
					</div>
				</div>
				<p class="help is-danger" v-if="showMustSelectUserError">
					{{ $t('team.edit.mustSelectUser') }}
				</p>
			</div>
			<table class="table has-actions is-striped is-hoverable is-fullwidth">
				<tbody>
				<tr :key="m.id" v-for="m in team?.members">
					<td>{{ getDisplayName(m) }}</td>
					<td>
						<template v-if="m.id === userInfo.id">
							<b class="is-success">You</b>
						</template>
					</td>
					<td class="type">
						<template v-if="m.admin">
								<span class="icon is-small">
									<icon icon="lock"/>
								</span>
							{{ $t('team.attributes.admin') }}
						</template>
						<template v-else>
								<span class="icon is-small">
									<icon icon="user"/>
								</span>
							{{ $t('team.attributes.member') }}
						</template>
					</td>
					<td class="actions" v-if="userIsAdmin">
						<x-button
							:loading="teamMemberService.loading"
							@click="() => toggleUserType(m)"
							class="mr-2"
							v-if="m.id !== userInfo.id"
						>
							{{ m.admin ? $t('team.edit.makeMember') : $t('team.edit.makeAdmin') }}
						</x-button>
						<x-button
							:loading="teamMemberService.loading"
							@click="() => {memberToDelete = m; showUserDeleteModal = true}"
							class="is-danger"
							v-if="m.id !== userInfo.id"
							icon="trash-alt"
						/>
					</td>
				</tr>
				</tbody>
			</table>
		</card>

		<x-button class="is-fullwidth is-danger" @click="showLeaveModal = true">
			{{ $t('team.edit.leave.title') }}
		</x-button>

		<!-- Leave team modal -->
		<modal
			v-if="showLeaveModal"
			@close="showLeaveModal = false"
			@submit="leave()"
		>
			<template #header><span>{{ $t('team.edit.leave.title') }}</span></template>

			<template #text>
				<p>{{ $t('team.edit.leave.text1') }}<br/>
					{{ $t('team.edit.leave.text2') }}</p>
			</template>
		</modal>

		<!-- Team delete modal -->
		<modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteTeam()"
		>
			<template #header><span>{{ $t('team.edit.delete.header') }}</span></template>

			<template #text>
				<p>{{ $t('team.edit.delete.text1') }}<br/>
					{{ $t('team.edit.delete.text2') }}</p>
			</template>
		</modal>

		<!-- User delete modal -->
		<modal
			:enabled="showUserDeleteModal"
			@close="showUserDeleteModal = false"
			@submit="deleteMember()"
		>
			<template #header><span>{{ $t('team.edit.deleteUser.header') }}</span></template>

			<template #text>
				<p>{{ $t('team.edit.deleteUser.text1') }}<br/>
					{{ $t('team.edit.deleteUser.text2') }}</p>
			</template>
		</modal>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'

import Editor from '@/components/input/AsyncEditor'
import Multiselect from '@/components/input/multiselect.vue'

import TeamService from '@/services/team'
import TeamMemberService from '@/services/teamMember'
import UserService from '@/services/user'

import {RIGHTS as Rights} from '@/constants/rights'

import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {getDisplayName} from '@/models/user'
import {useAuthStore} from '@/stores/auth'

import type {ITeam} from '@/modelTypes/ITeam'
import type {IUser} from '@/modelTypes/IUser'
import type {ITeamMember} from '@/modelTypes/ITeamMember'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const {t} = useI18n({useScope: 'global'})

const userIsAdmin = computed(() => {
	return (
		team.value &&
		team.value.maxRight &&
		team.value.maxRight > Rights.READ
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
	margin-bottom: 1rem;

	.content {
		padding: 0;
	}
}
</style>