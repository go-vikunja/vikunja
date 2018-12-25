<template>
	<div class="fullpage">
		<a class="close" @click="back()">
			<icon :icon="['far', 'times-circle']">
			</icon>
		</a>
		<h3>Create a new team</h3>
		<form @submit.prevent="newTeam" @keyup.esc="back()">
			<div class="field is-grouped">
				<p class="control is-expanded" v-bind:class="{ 'is-loading': loading}">
					<input v-focus class="input" v-bind:class="{ 'disabled': loading}" v-model="team.name" type="text" placeholder="The team's name goes here...">
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
		created() {
			this.$parent.setFullPage();
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
			back() {
				router.go(-1)
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
