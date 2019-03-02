<template>
	<div class="card">

		<header class="card-header">
			<p class="card-header-title">
				Teams with access to this {{typeString}}
			</p>
		</header>
		<div class="card-content content teams-list">
			<form @submit.prevent="addTeam()" class="add-team-form" v-if="userIsAdmin">
				<div class="field is-grouped">
					<p class="control has-icons-left is-expanded" v-bind:class="{ 'is-loading': this.teamService.loading}">
						<input class="input" v-bind:class="{ 'disabled': this.teamService.loading}" v-model.number="teamStuffModel.teamID" type="text" placeholder="Add a new team...">
						<span class="icon is-small is-left">
								<icon icon="users"/>
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
				<tr v-for="t in listTeams" :key="t.id">
					<td>
						<router-link :to="{name: 'editTeam', params: {id: t.id}}">
							{{t.name}}
						</router-link>
					</td>
					<td class="type">
						<template v-if="t.right === 2">
									<span class="icon is-small">
										<icon icon="lock"/>
									</span>
							Admin
						</template>
						<template v-else-if="t.right === 1">
									<span class="icon is-small">
										<icon icon="pen"/>
									</span>
							Write
						</template>
						<template v-else>
									<span class="icon is-small">
										<icon icon="users"/>
									</span>
							Read-only
						</template>
					</td>
					<td class="actions" v-if="userIsAdmin">
						<button @click="toggleTeamType(t.id, (t.right === 2))" class="button buttonright is-primary">
							Make
							<template v-if="t.right === 2">
								Member
							</template>
							<template v-else>
								Admin
							</template>
						</button>
						<button @click="teamToDelete = t.id; showTeamDeleteModal = true" class="button is-danger">
									<span class="icon is-small">
										<icon icon="trash-alt"/>
									</span>
						</button>
					</td>
				</tr>
				</tbody>
			</table>
		</div>

		<modal
				v-if="showTeamDeleteModal"
				@close="showTeamDeleteModal = false"
				v-on:submit="deleteTeam()">
			<span slot="header">Remove a team from the {{typeString}}</span>
			<p slot="text">Are you sure you want to remove this team from the {{typeString}}?<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
	import auth from '../../auth'
	import message from '../../message'
	import TeamNamespaceService from '../../services/teamNamespace'
	import TeamNamespaceModel from '../../models/teamNamespace'
	import TeamListModel from '../../models/teamList'
	import TeamListService from '../../services/teamList'
	
	export default {
		name: 'team',
		props: {
			type: '',
			id: 0,
			userIsAdmin: false,
		},
		data() {
			return {
				teamService: Object, // This team service is either a teamNamespaceService or a teamListService, depending on the type we are using
				teamStuffModel: Object,

				currentUser: auth.user.infos,
				typeString: '',
				listTeams: [],
				newTeam: {teamID: 0},
				showTeamDeleteModal: false,
				teamToDelete: 0,
			}
		},
		created() {
			if (this.type === 'list') {
				this.typeString = `list`
				this.teamService = new TeamListService()
				this.teamStuffModel = new TeamListModel({listID: this.id})
			} else if (this.type === 'namespace') {
				this.typeString = `namespace`
				this.teamService = new TeamNamespaceService()
				this.teamStuffModel = new TeamNamespaceModel({namespaceID: this.id})
			} else {
				throw new Error('Unknown type: ' + this.type)
			}

			this.loadTeams()
		},
		methods: {
			loadTeams() {
				this.teamService.getAll(this.teamStuffModel)
					.then(r => {
						this.$set(this, 'listTeams', r)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			deleteTeam() {
				this.teamService.delete(this.teamStuffModel)
					.then(() => {
						this.showTeamDeleteModal = false;
						message.success({message: 'The team was successfully deleted from the ' + this.typeString + '.'}, this)
						// FIXME: this should remove the team from the list instead of loading it again
						this.loadTeams()
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			addTeam(admin) {
				if(admin === null) {
					admin = false
				}
				this.teamStuffModel.right = 0
				if (admin) {
					this.teamStuffModel.right = 2
				}

				this.teamService.create(this.teamStuffModel)
					.then(() => {
						// FIXME: this should add the team to the list instead of loading it again
						this.loadTeams()
						message.success({message: 'The team was successfully added.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			toggleTeamType(teamid, current) {
				this.teamStuffModel.teamID = teamid
				this.teamStuffModel.right = 0
				if (!current) {
					this.teamStuffModel.right = 2
				}

				this.teamService.update(this.teamStuffModel)
					.then(() => {
						// FIXME: this should update the team in the list instead of loading it again
						this.loadTeams()
						message.success({message: 'The team right was successfully updated.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			}
		},
	}
</script>

<style lang="scss" scoped>
	.card{
		margin-bottom: 1rem;

		.add-team-form {
			margin: 1rem;
		}

		.table{
			border-top: 1px solid darken(#fff, 15%);

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

	.teams-list, .teams-namespace{
		padding: 0 !important;
	}
</style>