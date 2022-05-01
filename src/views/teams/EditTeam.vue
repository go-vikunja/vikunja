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
							:disabled="teamMemberService.loading || null"
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
					v-if="showError && team.name === ''"
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
			</div>
			<table class="table has-actions is-striped is-hoverable is-fullwidth">
				<tbody>
				<tr :key="m.id" v-for="m in team?.members">
					<td>{{ m.getDisplayName() }}</td>
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

		<!-- Team delete modal -->
		<transition name="modal">
			<modal
				@close="showDeleteModal = false"
				@submit="deleteTeam()"
				v-if="showDeleteModal"
			>
				<template #header><span>{{ $t('team.edit.delete.header') }}</span></template>

				<template #text>
					<p>{{ $t('team.edit.delete.text1') }}<br/>
						{{ $t('team.edit.delete.text2') }}</p>
				</template>
			</modal>
		</transition>
		<!-- User delete modal -->
		<transition name="modal">
			<modal
				@close="showUserDeleteModal = false"
				@submit="deleteMember()"
				v-if="showUserDeleteModal"
			>
				<template #header><span>{{ $t('team.edit.deleteUser.header') }}</span></template>

				<template #text>
					<p>{{ $t('team.edit.deleteUser.text1') }}<br/>
						{{ $t('team.edit.deleteUser.text2') }}</p>
				</template>
			</modal>
		</transition>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref} from 'vue'

import Editor from '@/components/input/AsyncEditor'
import {useStore} from 'vuex'

import TeamService from '../../services/team'
import TeamModel from '../../models/team'
import TeamMemberService from '../../services/teamMember'
import TeamMemberModel from '../../models/teamMember'
import UserModel from '../../models/user'
import UserService from '../../services/user'
import Rights from '../../models/constants/rights.json'

import Multiselect from '@/components/input/multiselect.vue'
import {useRoute, useRouter} from 'vue-router'
import {useTitle} from '@/composables/useTitle'
import {useI18n} from 'vue-i18n'
import {success} from '@/message'

const store = useStore()
const route = useRoute()
const router = useRouter()
const {t} = useI18n()

const userIsAdmin = computed(() => {
	return (
		team.value &&
		team.value.maxRight &&
		team.value.maxRight > Rights.READ
	)
})
const userInfo = computed(() => store.state.auth.info)

const teamService = ref<TeamService>(new TeamService())
const teamMemberService = ref<TeamMemberService>(new TeamMemberService())
const userService = ref<UserService>(new UserService())

const team = ref<TeamModel>()
const teamId = computed(() => route.params.id)
const memberToDelete = ref<TeamMemberModel>()
const newMember = ref<UserModel>()
const foundUsers = ref<UserModel[]>()

const showDeleteModal = ref(false)
const showUserDeleteModal = ref(false)
const showError = ref(false)

const title = ref('')

loadTeam()

async function loadTeam() {
	team.value = await teamService.value.get({id: teamId.value})
	title.value = t('team.edit.title', {team: team.value?.name})
	useTitle(() => title.value)
}

async function save() {
	if (team.value?.name === '') {
		showError.value = true
		return
	}
	showError.value = false

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
	await teamMemberService.value.create({
		teamId: teamId.value,
		username: newMember.value.username,
	})
	await loadTeam()
	success({message: t('team.edit.userAddedSuccess')})
}

async function toggleUserType(member) {
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
	foundUsers.value = users.filter((u: UserModel) => u.id !== userInfo.value.id)
}
</script>

<style lang="scss" scoped>
.card.is-fullwidth {
	margin-bottom: 1rem;

	.content {
		padding: 0;
	}
}

@include modal-transition();
</style>