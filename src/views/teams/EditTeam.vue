<template>
	<div class="loader-container is-max-width-desktop" v-bind:class="{ 'is-loading': teamService.loading}">
		<div class="card is-fullwidth" v-if="userIsAdmin">
			<header class="card-header">
				<p class="card-header-title">
					Edit Team
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="teamtext">Team Name</label>
							<div class="control">
								<input
										v-focus
										:class="{ 'disabled': teamMemberService.loading}"
										:disabled="teamMemberService.loading"
										class="input"
										type="text"
										id="teamtext"
										placeholder="The team text is here..."
										v-model="team.name"/>
							</div>
						</div>
						<p class="help is-danger" v-if="showError && team.name === ''">
							Please specify a name.
						</p>
						<div class="field">
							<label class="label" for="teamdescription">Description</label>
							<div class="control">
								<editor
										:class="{ 'disabled': teamService.loading}"
										:disabled="teamService.loading"
										placeholder="The teams description goes here..."
										id="teamdescription"
										v-model="team.description"
										:preview-is-default="false"
								/>
							</div>
						</div>
					</form>

					<div class="columns bigbuttons">
						<div class="column">
							<button @click="submit()" class="button is-primary is-fullwidth"
									:class="{ 'is-loading': teamService.loading}">
								Save
							</button>
						</div>
						<div class="column is-1">
							<button @click="showDeleteModal = true" class="button is-danger is-fullwidth"
									:class="{ 'is-loading': teamService.loading}">
								<span class="icon is-small">
									<icon icon="trash-alt"/>
								</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
		<div class="card is-fullwidth">
			<header class="card-header">
				<p class="card-header-title">
					Team Members
				</p>
			</header>
			<div class="card-content content team-members">
				<form @submit.prevent="addUser()" class="add-member-form" v-if="userIsAdmin">
					<div class="field is-grouped">
						<p
								class="control has-icons-left is-expanded"
								:class="{ 'is-loading': teamMemberService.loading}">
							<multiselect
									v-model="newMember"
									:options="foundUsers"
									:multiple="false"
									:searchable="true"
									:loading="userService.loading"
									:internal-search="true"
									@search-change="findUser"
									placeholder="Type to search..."
									:showNoOptions="false"
									label="username"
									track-by="id">
								<template slot="clear" slot-scope="props">
									<div
											class="multiselect__clear" v-if="newMember !== null && newMember.id !== 0"
											@mousedown.prevent.stop="clearAll(props.search)">
									</div>
								</template>
								<span slot="noResult">Oops! No user found. Consider changing the search query.</span>
							</multiselect>
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
							<template v-if="m.id === userInfo.id">
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
							<button @click="toggleUserType(m)" class="button buttonright is-primary"
									:class="{'is-loading': teamMemberService.loading}"
									v-if="m.id !== userInfo.id">
								Make
								<template v-if="!m.admin">
									Admin
								</template>
								<template v-else>
									Member
								</template>
							</button>
							<button @click="() => {member = m; showUserDeleteModal = true}" class="button is-danger"
									:class="{'is-loading': teamMemberService.loading}"
									v-if="m.id !== userInfo.id">
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
				@submit="deleteTeam()">
			<span slot="header">Delete the team</span>
			<p slot="text">Are you sure you want to delete this team and all of its members?<br/>
				All team members will loose access to lists and namespaces shared with this team.<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
		<!-- User delete modal -->
		<modal
				v-if="showUserDeleteModal"
				@close="showUserDeleteModal = false"
				@submit="deleteUser()">
			<span slot="header">Remove a user from the team</span>
			<p slot="text">Are you sure you want to remove this user from the team?<br/>
				They will loose access to all lists and namespaces this team has access to.<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
	import router from '../../router'
	import {mapState} from 'vuex'

	import TeamService from '../../services/team'
	import TeamModel from '../../models/team'
	import TeamMemberService from '../../services/teamMember'
	import TeamMemberModel from '../../models/teamMember'
	import UserModel from '../../models/user'
	import UserService from '../../services/user'
	import LoadingComponent from '../../components/misc/loading'
	import ErrorComponent from '../../components/misc/error'
	import Rights from '../../models/rights.json'

	export default {
		name: 'EditTeam',
		data() {
			return {
				teamService: TeamService,
				teamMemberService: TeamMemberService,
				team: TeamModel,
				teamId: this.$route.params.id,
				member: TeamMemberModel,

				showDeleteModal: false,
				showUserDeleteModal: false,

				newMember: UserModel,
				foundUsers: [],
				userService: UserService,

				showError: false,
			}
		},
		components: {
			multiselect: () => ({
				component: import(/* webpackPrefetch: true *//* webpackChunkName: "multiselect" */ 'vue-multiselect'),
				loading: LoadingComponent,
				error: ErrorComponent,
				timeout: 60000,
			}),
			editor: () => ({
				component: import(/* webpackPrefetch: true *//* webpackChunkName: "editor" */ '../../components/input/editor'),
				loading: LoadingComponent,
				error: ErrorComponent,
				timeout: 60000,
			}),
		},
		created() {
			this.teamService = new TeamService()
			this.teamMemberService = new TeamMemberService()
			this.userService = new UserService()
			this.loadTeam()
		},
		watch: {
			// call again the method if the route changes
			'$route': 'loadTeam',
		},
		computed: {
			userIsAdmin() {
				return this.team && this.team.maxRight && this.team.maxRight > Rights.READ
			},
			...mapState({
				userInfo: state => state.auth.info,
			}),
		},
		methods: {
			loadTeam() {
				this.team = new TeamModel({id: this.teamId})
				this.teamService.get(this.team)
					.then(response => {
						this.$set(this, 'team', response)
						this.setTitle(`Edit Team ${this.team.name}`)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			submit() {
				if (this.team.name === '') {
					this.showError = true
					return
				}
				this.showError = false

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
						router.push({name: 'teams.index'})
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
				const newMember = new TeamMemberModel({
					teamId: this.teamId,
					username: this.newMember.username,
				})
				this.teamMemberService.create(newMember)
					.then(() => {
						this.loadTeam()
						this.success({message: 'The team member was successfully added.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			toggleUserType(member) {
				member.admin = !member.admin
				this.teamMemberService.update(member)
					.then(r => {
						for (const tm in this.team.members) {
							if (this.team.members[tm].id === member.id) {
								this.$set(this.team.members[tm], 'admin', r.admin)
								break
							}
						}
						this.success({message: 'The team member was successfully made ' + (member.admin ? 'admin' : 'member') + '.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			findUser(query) {
				if (query === '') {
					this.$set(this, 'foundUsers', [])
					return
				}

				this.userService.getAll({}, {s: query})
					.then(response => {
						this.$set(this, 'foundUsers', response)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			clearAll() {
				this.$set(this, 'foundUsers', [])
			},
		},
	}
</script>

<style lang="scss" scoped>
	.card {
		margin-bottom: 1rem;

		.add-member-form {
			margin: 1rem;
		}
	}

	.team-members {
		padding: 0;

		.table {
			border-top: 0;
		}
	}
</style>