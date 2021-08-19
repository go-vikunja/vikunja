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
				<tr :key="m.id" v-for="m in team.members">
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
							@click="() => {member = m; showUserDeleteModal = true}"
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
				@submit="deleteUser()"
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

<script>
import {mapState} from 'vuex'

import TeamService from '../../services/team'
import TeamModel from '../../models/team'
import TeamMemberService from '../../services/teamMember'
import TeamMemberModel from '../../models/teamMember'
import UserModel from '../../models/user'
import UserService from '../../services/user'
import Rights from '../../models/constants/rights.json'

import LoadingComponent from '../../components/misc/loading'
import ErrorComponent from '../../components/misc/error'

import Multiselect from '@/components/input/multiselect.vue'

export default {
	name: 'EditTeam',
	data() {
		return {
			teamService: new TeamService(),
			teamMemberService: new TeamMemberService(),
			team: TeamModel,
			teamId: this.$route.params.id,
			member: TeamMemberModel,

			showDeleteModal: false,
			showUserDeleteModal: false,

			newMember: UserModel,
			foundUsers: [],
			userService: new UserService(),

			showError: false,
			title: '',
		}
	},
	components: {
		Multiselect,
		editor: () => ({
			component: import('../../components/input/editor'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	watch: {
		// call again the method if the route changes
		'$route': {
			handler: 'loadTeam',
			deep: true,
			immediate: true,
		},
	},
	computed: {
		userIsAdmin() {
			return (
				this.team &&
				this.team.maxRight &&
				this.team.maxRight > Rights.READ
			)
		},
		...mapState({
			userInfo: (state) => state.auth.info,
		}),
	},
	methods: {
		loadTeam() {
			this.team = new TeamModel({id: this.teamId})
			this.teamService
				.get(this.team)
				.then((response) => {
					this.team = response
					this.title = this.$t('team.edit.title', {team: this.team.name})
					this.setTitle(this.title)
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		save() {
			if (this.team.name === '') {
				this.showError = true
				return
			}
			this.showError = false

			this.teamService
				.update(this.team)
				.then((response) => {
					this.team = response
					this.$message.success({message: this.$t('team.edit.success')})
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		deleteTeam() {
			this.teamService
				.delete(this.team)
				.then(() => {
					this.$message.success({message: this.$t('team.edit.delete.success')})
					this.$router.push({name: 'teams.index'})
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		deleteUser() {
			this.teamMemberService
				.delete(this.member)
				.then(() => {
					this.$message.success({message: this.$t('team.edit.deleteUser.success')})
					this.loadTeam()
				})
				.catch((e) => {
					this.$message.error(e)
				})
				.finally(() => {
					this.showUserDeleteModal = false
				})
		},
		addUser() {
			const newMember = new TeamMemberModel({
				teamId: this.teamId,
				username: this.newMember.username,
			})
			this.teamMemberService
				.create(newMember)
				.then(() => {
					this.loadTeam()
					this.$message.success({message: this.$t('team.edit.userAddedSuccess')})
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		toggleUserType(member) {
			member.admin = !member.admin
			member.teamId = this.teamId
			this.teamMemberService
				.update(member)
				.then((r) => {
					for (const tm in this.team.members) {
						if (this.team.members[tm].id === member.id) {
							this.team.members[tm].admin = r.admin
							break
						}
					}
					this.$message.success({
						message: member.admin ?
							this.$t('team.edit.madeAdmin') :
							this.$t('team.edit.madeMember'),
					})
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		findUser(query) {
			if (query === '') {
				this.clearAll()
				return
			}

			this.userService
				.getAll({}, {s: query})
				.then((response) => {
					this.foundUsers = response
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		clearAll() {
			this.foundUsers = []
		},
	},
}
</script>
