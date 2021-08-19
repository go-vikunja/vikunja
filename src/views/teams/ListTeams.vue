<template>
	<div class="content loader-container is-max-width-desktop" :class="{ 'is-loading': teamService.loading}">
		<x-button
			:to="{name:'teams.create'}"
			class="is-pulled-right"
			icon="plus"
		>
			{{ $t('team.create.title') }}
		</x-button>

		<h1>{{ $t('team.title') }}</h1>
		<ul class="teams box" v-if="teams.length > 0">
			<li :key="t.id" v-for="t in teams">
				<router-link :to="{name: 'teams.edit', params: {id: t.id}}">
					{{ t.name }}
				</router-link>
			</li>
		</ul>
		<p v-else-if="!teamService.loading" class="has-text-centered has-text-grey is-italic">
			{{ $t('team.noTeams') }}
			<router-link :to="{name: 'teams.create'}">
				{{ $t('team.create.title') }}.
			</router-link>
		</p>
	</div>
</template>

<script>
import TeamService from '../../services/team'

export default {
	name: 'ListTeams',
	data() {
		return {
			teamService: new TeamService(),
			teams: [],
		}
	},
	created() {
		this.loadTeams()
	},
	mounted() {
		this.setTitle(this.$t('team.title'))
	},
	methods: {
		loadTeams() {
			this.teamService.getAll()
				.then(response => {
					this.teams = response
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
	},
}
</script>
