<template>
	<div class="fullpage">
		<a @click="back()" class="close">
			<icon :icon="['far', 'times-circle']">
			</icon>
		</a>
		<h3>Create a new team</h3>
		<form @keyup.esc="back()" @submit.prevent="newTeam">
			<div class="field is-grouped">
				<p class="control is-expanded" v-bind:class="{ 'is-loading': teamService.loading}">
					<input
						:class="{ 'disabled': teamService.loading}"
						class="input"
						placeholder="The team's name goes here..." type="text"
						v-focus
						v-model="team.name"/>
				</p>
				<p class="control">
					<button class="button is-primary has-no-shadow" type="submit">
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
import {IS_FULLPAGE} from '@/store/mutation-types'

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
	mounted() {
		this.setTitle('Create a new Team')
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
	},
}
</script>
