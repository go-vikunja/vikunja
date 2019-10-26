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
					<input v-focus class="input" v-bind:class="{ 'disabled': teamService.loading}" v-model="team.name" type="text" placeholder="The team's name goes here...">
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
		</form>
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'
	import message from '../../message'
	import TeamModel from '../../models/team'
	import TeamService from '../../services/team'

	export default {
		name: "NewTeam",
		data() {
			return {
				teamService: TeamService,
				team: TeamModel,
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
			this.team = new TeamModel()
			this.$parent.setFullPage();
		},
		methods: {
			newTeam() {
				this.teamService.create(this.team)
					.then(response => {
						router.push({name:'editTeam', params:{id: response.id}})
						message.success({message: 'The team was successfully created.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			back() {
				router.go(-1)
			},
		}
	}
</script>
