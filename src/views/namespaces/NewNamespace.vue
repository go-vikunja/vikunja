<template>
	<create-edit
		:title="$t('namespace.create.title')"
		@create="newNamespace()"
		:create-disabled="namespace.title === ''"
	>
		<div class="field">
			<label class="label" for="namespaceTitle">{{ $t('namespace.attributes.title') }}</label>
			<div
				class="control is-expanded"
				:class="{ 'is-loading': namespaceService.loading }"
			>
				<input
					@keyup.enter="newNamespace()"
					@keyup.esc="back()"
					class="input"
					:placeholder="$t('namespace.attributes.titlePlaceholder')"
					type="text"
					:class="{ disabled: namespaceService.loading }"
					v-focus
					v-model="namespace.title"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && namespace.title === ''">
			{{ $t('namespace.create.titleRequired') }}
		</p>
		<div class="field">
			<label class="label">{{ $t('namespace.attributes.color') }}</label>
			<div class="control">
				<color-picker v-model="namespace.hexColor" />
			</div>
		</div>
		<p
			class="is-small has-text-centered"
			v-tooltip.bottom="$t('namespace.create.explanation')"
		>
			{{ $t('namespace.create.tooltip') }}
		</p>
	</create-edit>
</template>

<script>
import NamespaceModel from '../../models/namespace'
import NamespaceService from '../../services/namespace'
import CreateEdit from '@/components/misc/create-edit.vue'
import ColorPicker from '../../components/input/colorPicker'

export default {
	name: 'NewNamespace',
	data() {
		return {
			showError: false,
			namespace: new NamespaceModel(),
			namespaceService: new NamespaceService(),
		}
	},
	components: {
		ColorPicker,
		CreateEdit,
	},
	mounted() {
		this.setTitle(this.$t('namespace.create.title'))
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
					this.$message.success({message: this.$t('namespace.create.success') })
					this.$router.back()
				})
		},
	},
}
</script>
