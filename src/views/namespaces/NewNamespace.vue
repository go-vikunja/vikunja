<template>
	<create
		title="Create a new namespace"
		@create="newNamespace()"
		:create-disabled="namespace.title === ''"
	>
		<div class="field">
			<label class="label" for="namespaceTitle">Namespace Title</label>
			<div
				class="control is-expanded"
				:class="{ 'is-loading': namespaceService.loading }"
			>
				<input
					@keyup.enter="newNamespace()"
					@keyup.esc="back()"
					class="input"
					placeholder="The namespace's name goes here..."
					type="text"
					:class="{ disabled: namespaceService.loading }"
					v-focus
					v-model="namespace.title"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && namespace.title === ''">
			Please specify a title.
		</p>
		<div class="field">
			<label class="label">Color</label>
			<div class="control">
				<color-picker v-model="namespace.hexColor" />
			</div>
		</div>
		<p
			class="is-small has-text-centered"
			v-tooltip.bottom="
				'A namespace is a collection of lists you can share and use to organize your lists with. In fact, every list belongs to a namepace.'
			"
		>
			What's a namespace?
		</p>
	</create>
</template>

<script>
import NamespaceModel from '../../models/namespace'
import NamespaceService from '../../services/namespace'
import Create from '@/components/misc/create'
import ColorPicker from '../../components/input/colorPicker'

export default {
	name: 'NewNamespace',
	data() {
		return {
			showError: false,
			namespace: NamespaceModel,
			namespaceService: NamespaceService,
		}
	},
	components: {
		ColorPicker,
		Create,
	},
	created() {
		this.namespace = new NamespaceModel()
		this.namespaceService = new NamespaceService()
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

			this.namespaceService
				.create(this.namespace)
				.then((r) => {
					this.$store.commit('namespaces/addNamespace', r)
					this.success(
						{ message: 'The namespace was successfully created.' },
						this
					)
					this.$router.back()
				})
				.catch((e) => {
					this.error(e, this)
				})
		},
	},
}
</script>
