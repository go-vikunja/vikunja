<template>
	<div class="content">
		<h3>Create a new team</h3>
		<form @submit.prevent="newTeam">
			<div class="field is-grouped">
				<p class="control has-icons-left is-expanded" v-bind:class="{ 'is-loading': loading}">
					<input class="input" v-bind:class="{ 'disabled': loading}" v-model="team.name" type="text" placeholder="The team's name goes here...">
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
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'
	import {HTTP} from '../../http-common'
	import message from '../../message'

	export default {
		name: "NewTeam",
		data() {
			return {
				team: {title: ''},
				error: '',
				loading: false
			}
		},
		beforeMount() {
			// Check if the user is already logged in, if so, redirect him to the homepage
			if (!auth.user.authenticated) {
				router.push({name: 'home'})
			}
		},
		methods: {
			newTeam() {
				const cancel = message.setLoading(this)

				HTTP.put(`teams`, this.team, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						router.push({name:'editTeam', params:{id: response.data.id}})
						this.handleSuccess({message: 'The team was successfully created.'})
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
		}
	}
</script>

<style scoped>

</style>