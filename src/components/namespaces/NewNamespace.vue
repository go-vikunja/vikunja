<template>
	<div class="fullpage">
		<a class="close" @click="back()">
			<icon :icon="['far', 'times-circle']">
			</icon>
		</a>
		<h3>Create a new namespace</h3>
		<form @submit.prevent="newNamespace" @keyup.esc="back()">
			<div class="field is-grouped">
				<p class="control is-expanded" v-bind:class="{ 'is-loading': loading}">
					<input v-focus class="input" v-bind:class="{ 'disabled': loading}" v-model="namespace.name" type="text" placeholder="The namespace's name goes here...">
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
		<p class="small" v-tooltip.bottom="'A namespace is a collection of lists you can share and use to organize your lists with.<br/>In fact, every list belongs to a namepace.'">What's a namespace?</p>
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'
	import {HTTP} from '../../http-common'
	import message from '../../message'

	export default {
		name: "NewNamespace",
		data() {
			return {
				namespace: {title: ''},
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
			newNamespace() {
				const cancel = message.setLoading(this)

				HTTP.put(`namespaces`, this.namespace, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(() => {
						this.$parent.loadNamespaces()
						this.handleSuccess({message: 'The namespace was successfully created.'})
						cancel()
						router.push({name: 'home'})
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
