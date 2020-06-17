<template>
	<div class="fullpage">
		<a class="close" @click="back()">
			<icon :icon="['far', 'times-circle']">
			</icon>
		</a>
		<h3>Create a new team</h3>
		<form @submit.prevent="newTeam" @keyup.esc="back()">
			<div class="field is-grouped">
				<p class="control is-expanded" v-bind:class="{ 'is-loading': teamService.loading}">
					<input
							v-focus
							class="input"
							:class="{ 'disabled': teamService.loading}" v-model="team.name"
							type="text"
							placeholder="The team's name goes here..."/>
				</p>
				<p class="control">
					<button type="submit" class="button is-success noshadow">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
						Add
					</button>
				</p>
			</div>
			<p class="help is-danger" v-if="showError && team.name === ''">
				Please specify a name.
			</p>
		</form>
	</div>
</template>

<script>
	import router from '../../router'
	import TeamModel from '../../models/team'
	import TeamService from '../../services/team'
	import {IS_FULLPAGE} from '../../store/mutation-types'

	export default {
		name: 'NewTeam',
		data() {
			return {
				teamService: TeamService,
				team: TeamModel,
				showError: false,
			}
		},
		created() {
			this.teamService = new TeamService()
			this.team = new TeamModel()
			this.$store.commit(IS_FULLPAGE, true)
		},
		methods: {
			newTeam() {

				if (this.team.name === '') {
					this.showError = true
					return
				}
				this.showError = false

				this.teamService.create(this.team)
					.then(response => {
						router.push({name: 'teams.edit', params: {id: response.id}})
						this.success({message: 'The team was successfully created.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			back() {
				router.go(-1)
			},
		}
	}
</script>
