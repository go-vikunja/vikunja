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
					<p class="control has-icons-left is-expanded" v-bind:class="{ 'is-loading': loading}">
						<input class="input" v-bind:class="{ 'disabled': loading}" v-model.number="newTeam.team_id" type="text" placeholder="Add a new team...">
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
	import {HTTP} from '../../http-common'
	import auth from '../../auth'
	import message from '../../message'
	
	export default {
		name: 'team',
		props: {
			type: '',
			id: 0,
			userIsAdmin: false,
		},
		data() {
			return {
				loading: false,
				currentUser: auth.user.infos,
				typeString: '',
				listTeams: [],
				newTeam: {team_id: 0},
				showTeamDeleteModal: false,
				teamToDelete: 0,
			}
		},
		created() {
			if (this.type === 'list') {
				this.typeString = `list`
			} else if (this.type === 'namespace') {
				this.typeString = `namespace`
			} else {
				throw new Error('Unknown type: ' + this.type)
			}

			this.loadTeams()
		},
		methods: {
			loadTeams() {
				const cancel = message.setLoading(this)
				HTTP.get(this.typeString + `s/` + this.id + `/teams`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						this.$set(this, 'listTeams', response.data)
						cancel()
					})
					.catch(e => {
						cancel()
						this.handleError(e)
					})
			},
			deleteTeam() {
				const cancel = message.setLoading(this)
				HTTP.delete(this.typeString + `s/` + this.id + `/teams/` + this.teamToDelete, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.showTeamDeleteModal = false;
						this.handleSuccess({message: 'The team was successfully deleted from the ' + this.typeString + '.'})
						this.loadTeams()
						cancel()
					})
					.catch(e => {
						cancel()
						this.handleError(e)
					})
			},
			addTeam(admin) {
				const cancel = message.setLoading(this)
				if(admin === null) {
					admin = false
				}
				this.newTeam.right = 0
				if (admin) {
					this.newTeam.right = 2
				}

				HTTP.put(this.typeString + `s/` + this.id + `/teams`, this.newTeam, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.loadTeams()
						this.handleSuccess({message: 'The team was successfully added.'})
						cancel()
					})
					.catch(e => {
						cancel()
						this.handleError(e)
					})
			},
			toggleTeamType(teamid, current) {
				const cancel = message.setLoading(this)
				let right = 0
				if (!current) {
					right = 2
				}

				HTTP.post(this.typeString + `s/` + this.id + `/teams/` + teamid, {right: right}, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.loadTeams()
						this.handleSuccess({message: 'The team right was successfully updated.'})
						cancel()
					})
					.catch(e => {
						cancel()
						this.handleError(e)
					})
			},
			handleError(e) {
				message.error(e, this)
			},
			handleSuccess(e) {
				message.success(e, this)
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