<template>
	<div class="loader-container" v-bind:class="{ 'is-loading': teamService.loading}">
		<div class="card" v-if="userIsAdmin">
			<header class="card-header">
				<p class="card-header-title">
					Edit Team
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form  @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="teamtext">Team Name</label>
							<div class="control">
								<input v-focus :class="{ 'disabled': teamMemberService.loading}" :disabled="teamMemberService.loading" class="input" type="text" id="teamtext" placeholder="The team text is here..." v-model="team.name">
							</div>
						</div>
						<div class="field">
							<label class="label" for="teamdescription">Description</label>
							<div class="control">
								<textarea :class="{ 'disabled': teamService.loading}" :disabled="teamService.loading" class="textarea" placeholder="The teams description goes here..." id="teamdescription" v-model="team.description"></textarea>
							</div>
						</div>
					</form>

					<div class="columns bigbuttons">
						<div class="column">
							<button @click="submit()" class="button is-success is-fullwidth" :class="{ 'is-loading': teamService.loading}">
								Save
							</button>
						</div>
						<div class="column is-1">
							<button @click="showDeleteModal = true" class="button is-danger is-fullwidth" :class="{ 'is-loading': teamService.loading}">
								<span class="icon is-small">
									<icon icon="trash-alt"/>
								</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
		<div class="card">

			<header class="card-header">
				<p class="card-header-title">
					Team Members
				</p>
			</header>
			<div class="card-content content team-members">
				<form @submit.prevent="addUser()" class="add-member-form" v-if="userIsAdmin">
					<div class="field is-grouped">
						<p class="control has-icons-left is-expanded" v-bind:class="{ 'is-loading': teamMemberService.loading}">
							<input class="input" v-bind:class="{ 'disabled': teamMemberService.loading}" v-model.number="member.id" type="text" placeholder="Add a new user...">
							<span class="icon is-small is-left">
								<icon icon="user"/>
							</span>
						</p>
						<p class="control">
							<button type="submit" class="button is-success">
								<span class="icon is-small">
									<icon icon="plus"/>
								</span>
								Add
							</button>
						</p>
					</div>
				</form>
				<table class="table is-striped is-hoverable is-fullwidth">
					<tbody>
						<tr v-for="m in team.members" :key="m.id">
							<td>{{m.username}}</td>
							<td>
								<template v-if="m.id === user.infos.id">
									<b class="is-success">You</b>
								</template>
							</td>
							<td class="type">
								<template v-if="m.admin">
									<span class="icon is-small">
										<icon icon="lock"/>
									</span>
									Admin
								</template>
								<template v-else>
									<span class="icon is-small">
										<icon icon="user"/>
									</span>
									Member
								</template>
							</td>
							<td class="actions" v-if="userIsAdmin">
								<button @click="toggleUserType(m)" class="button buttonright is-primary" v-if="m.id !== user.infos.id">
									Make
									<template v-if="!m.admin">
										Admin
									</template>
									<template v-else>
										Member
									</template>
								</button>
								<button @click="member = m; showUserDeleteModal = true" class="button is-danger" v-if="m.id !== user.infos.id">
									<span class="icon is-small">
										<icon icon="trash-alt"/>
									</span>
								</button>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>

		<!-- Team delete modal -->
		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				v-on:submit="deleteTeam()">
			<span slot="header">Delete the team</span>
			<p slot="text">Are you sure you want to delete this team and all of its members?<br/>
				All team members will loose access to lists and namespaces shared with this team.<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
		<!-- User delete modal -->
		<modal
				v-if="showUserDeleteModal"
				@close="showUserDeleteModal = false"
				v-on:submit="deleteUser()">
			<span slot="header">Remove a user from the team</span>
			<p slot="text">Are you sure you want to remove this user from the team?<br/>
				He will loose access to all lists and namespaces this team has access to.<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'

	import TeamService from '../../services/team'
	import TeamModel from '../../models/team'
	import TeamMemberService from '../../services/teamMember'
	import TeamMemberModel from '../../models/teamMember'
	
	export default {
		name: "EditTeam",
		data() {
			return {
				teamService: TeamService,
				teamMemberService: TeamMemberService,
				team: TeamModel,
				member: TeamMemberModel,

				showDeleteModal: false,
				showUserDeleteModal: false,
				user: auth.user,
				userIsAdmin: false,
			}
		},
		beforeMount() {
			// Check if the user is already logged in, if so, redirect him to the homepage
			if (!auth.user.authenticated) {
				router.push({name: 'home'})
			}
		},
		created() {
			this.teamService = new TeamService()
			this.teamMemberService = new TeamMemberService()
			this.loadTeam()
		},
		watch: {
			// call again the method if the route changes
			'$route': 'loadTeam'
		},
		methods: {
			loadTeam() {
				this.member = new TeamMemberModel({teamID: this.$route.params.id})
				this.team = new TeamModel({id: this.$route.params.id})
				this.teamService.get(this.team)
					.then(response => {
						this.$set(this, 'team', response)
						let members = response.members
						for (const m in members) {
							members[m].teamID = this.$route.params.id
							if (members[m].id === this.user.infos.id && members[m].admin) {
								this.userIsAdmin = true
							}
						}
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			submit() {
				this.teamService.update(this.team)
					.then(response => {
						this.team = response
						this.success({message: 'The team was successfully updated.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			deleteTeam() {
				this.teamService.delete(this.team)
					.then(() => {
						this.success({message: 'The team was successfully deleted.'}, this)
						router.push({name: 'listTeams'})
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			deleteUser() {
				this.teamMemberService.delete(this.member)
					.then(() => {
						this.success({message: 'The user was successfully deleted from the team.'}, this)
						this.loadTeam()
					})
					.catch(e => {
						this.error(e, this)
					})
					.finally(() => {
						this.showUserDeleteModal = false
					})
			},
			addUser() {
				this.teamMemberService.create(this.member)
					.then(() => {
						this.loadTeam()
						this.success({message: 'The team member was successfully added.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			toggleUserType(member) {
				this.member = member
				this.member.admin = !member.admin
				this.deleteUser()
				this.addUser()
			}
		}
	}
</script>

<style lang="scss" scoped>
	.card{
		margin-bottom: 1rem;

		.add-member-form {
			margin: 1rem;
		}

		.table{
			border-top: 1px solid darken(#fff, 15%);
			border-radius: 4px;
			overflow: hidden;

			td{
				vertical-align: middle;
			}

			td.type, td.actions{
				width: 200px;
			}

			td.actions{
				text-align: right;
			}
		}
	}

	.team-members{
		padding: 0;
	}
</style>