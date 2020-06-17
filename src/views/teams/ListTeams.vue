<template>
	<div class="content loader-container" v-bind:class="{ 'is-loading': teamService.loading}">
		<router-link :to="{name:'teams.create'}" class="button is-success button-right" >
			<span class="icon is-small">
				<icon icon="plus"/>
			</span>
			New Team
		</router-link>
		<h1>Teams</h1>
		<ul class="teams box">
			<li v-for="t in teams" :key="t.id">
				<router-link :to="{name: 'teams.edit', params: {id: t.id}}">
					{{t.name}}
				</router-link>
			</li>
		</ul>
	</div>
</template>

<script>
	import TeamService from '../../services/team'
	
	export default {
		name: "ListTeams",
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
		}
	}
</script>
