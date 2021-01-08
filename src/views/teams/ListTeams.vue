<template>
	<div class="content loader-container is-max-width-desktop" v-bind:class="{ 'is-loading': teamService.loading}">
		<router-link :to="{name:'teams.create'}" class="button is-primary button-right">
			<span class="icon is-small">
				<icon icon="plus"/>
			</span>
			New Team
		</router-link>

		<h1>Teams</h1>
		<ul class="teams box" v-if="teams.length > 0">
			<li :key="t.id" v-for="t in teams">
				<router-link :to="{name: 'teams.edit', params: {id: t.id}}">
					{{ t.name }}
				</router-link>
			</li>
		</ul>
		<p v-else class="has-text-centered has-text-grey">You are currently not part of any teams.</p>
	</div>
</template>

<script>
import TeamService from '../../services/team'

export default {
	name: 'ListTeams',
	data() {
		return {
			teamService: TeamService,
			teams: [],
		}
	},
	created() {
		this.teamService = new TeamService()
		this.loadTeams()
	},
	mounted() {
		this.setTitle('Teams')
	},
	methods: {
		loadTeams() {
			this.teamService.getAll()
				.then(response => {
					this.$set(this, 'teams', response)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
	},
}
</script>
