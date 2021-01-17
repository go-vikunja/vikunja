<template>
	<div class="fullpage">
		<a @click="back()" class="close">
			<icon :icon="['far', 'times-circle']">
			</icon>
		</a>
		<h3>Create a new namespace</h3>
		<div class="field is-grouped">
			<p class="control is-expanded" v-bind:class="{ 'is-loading': namespaceService.loading}">
				<input
					@keyup.enter="newNamespace()"
					@keyup.esc="back()"
					class="input"
					placeholder="The namespace's name goes here..."
					type="text"
					:class="{ 'disabled': namespaceService.loading}"
					v-focus
					v-model="namespace.title"/>
			</p>
			<p class="control">
				<button :disabled="namespace.title === ''" @click="newNamespace()" class="button is-primary has-no-shadow">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
					Add
				</button>
			</p>
		</div>
		<p class="help is-danger" v-if="showError && namespace.title === ''">
			Please specify a title.
		</p>
		<p
			class="small"
			v-tooltip.bottom="'A namespace is a collection of lists you can share and use to organize your lists with. In fact, every list belongs to a namepace.'">
			What's a namespace?</p>
	</div>
</template>

<script>
import router from '../../router'
import NamespaceModel from '../../models/namespace'
import NamespaceService from '../../services/namespace'
import {IS_FULLPAGE} from '@/store/mutation-types'

export default {
	name: 'NewNamespace',
	data() {
		return {
			showError: false,
			namespace: NamespaceModel,
			namespaceService: NamespaceService,
		}
	},
	created() {
		this.namespace = new NamespaceModel()
		this.namespaceService = new NamespaceService()
		this.$store.commit(IS_FULLPAGE, true)
	},
	mounted() {
		this.setTitle('Create a new namespace')
	},
	methods: {
		newNamespace() {
			if (this.namespace.title === '') {
				this.showError = true
				return
			}
			this.showError = false

			this.namespaceService.create(this.namespace)
				.then(r => {
					this.$store.commit('namespaces/addNamespace', r)
					this.success({message: 'The namespace was successfully created.'}, this)
					router.back()
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
