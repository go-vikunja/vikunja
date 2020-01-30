<template>
	<div class="fullpage">
		<a class="close" @click="back()">
			<icon :icon="['far', 'times-circle']">
			</icon>
		</a>
		<h3>Create a new namespace</h3>
		<form @submit.prevent="newNamespace" @keyup.esc="back()">
			<div class="field is-grouped">
				<p class="control is-expanded" v-bind:class="{ 'is-loading': namespaceService.loading}">
					<input v-focus class="input" v-bind:class="{ 'disabled': namespaceService.loading}" v-model="namespace.name" type="text" placeholder="The namespace's name goes here...">
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
	import NamespaceModel from "../../models/namespace";
	import NamespaceService from "../../services/namespace";

	export default {
		name: "NewNamespace",
		data() {
			return {
				namespace: NamespaceModel,
				namespaceService: NamespaceService,
			}
		},
		beforeMount() {
			// Check if the user is already logged in, if so, redirect him to the homepage
			if (!auth.user.authenticated) {
				router.push({name: 'home'})
			}
		},
		created() {
			this.namespace = new NamespaceModel()
			this.namespaceService = new NamespaceService()
			this.$parent.setFullPage();
		},
		methods: {
			newNamespace() {
				this.namespaceService.create(this.namespace)
					.then(() => {
						this.$parent.loadNamespaces()
						this.success({message: 'The namespace was successfully created.'}, this)
						router.push({name: 'home'})
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			back() {
				router.go(-1)
			}
		}
	}
</script>
