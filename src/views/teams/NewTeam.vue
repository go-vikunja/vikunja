<template>
	<create-edit
		title="Create a new team"
		@create="newTeam()"
		:create-disabled="team.name === ''"
	>
		<div class="field">
			<label class="label" for="teamName">Team Name</label>
			<div
				class="control is-expanded"
				:class="{ 'is-loading': teamService.loading }"
			>
				<input
					:class="{ 'disabled': teamService.loading }"
					class="input"
					id="teamName"
					placeholder="The team's name goes here..."
					type="text"
					v-focus
					v-model="team.name"
					@keyup.enter="newTeam"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && team.name === ''">
			Please specify a name.
		</p>
	</create-edit>
</template>

<script>
import TeamModel from '../../models/team'
import TeamService from '../../services/team'
import CreateEdit from '@/components/misc/create-edit'

export default {
	name: 'NewTeam',
	data() {
		return {
			teamService: TeamService,
			team: TeamModel,
			showError: false,
		}
	},
	components: {
		CreateEdit,
	},
	created() {
		this.teamService = new TeamService()
		this.team = new TeamModel()
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

			this.teamService
				.create(this.team)
				.then((response) => {
					this.$router.push({
						name: 'teams.edit',
						params: { id: response.id },
					})
					this.success(
						{ message: 'The team was successfully created.' },
						this
					)
				})
				.catch((e) => {
					this.error(e, this)
				})
		},
	},
}
</script>
