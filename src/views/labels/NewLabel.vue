<template>
	<create
		title="Create a new label"
		@create="newLabel()"
		:create-disabled="label.title === ''"
	>
		<div class="field">
			<label class="label" for="labelTitle">Label Title</label>
			<div
				class="control is-expanded"
				:class="{ 'is-loading': labelService.loading }"
			>
				<input
					:class="{ disabled: labelService.loading }"
					class="input"
					placeholder="The label title goes here..."
					type="text"
					id="labelTitle"
					v-focus
					v-model="label.title"
					@keyup.enter="newLabel()"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && label.title === ''">
			Please specify a title.
		</p>
		<div class="field">
			<label class="label">Color</label>
			<div class="control">
				<color-picker v-model="label.hexColor" />
			</div>
		</div>
	</create>
</template>

<script>
import labelModel from '../../models/label'
import labelService from '../../services/label'
import LabelModel from '../../models/label'
import LabelService from '../../services/label'
import Create from '@/components/misc/create'
import ColorPicker from '../../components/input/colorPicker'

export default {
	name: 'NewLabel',
	data() {
		return {
			labelService: labelService,
			label: labelModel,
			showError: false,
		}
	},
	components: {
		Create,
		ColorPicker,
	},
	created() {
		this.labelService = new LabelService()
		this.label = new LabelModel()
	},
	mounted() {
		this.setTitle('Create a new label')
	},
	methods: {
		newLabel() {
			if (this.label.title === '') {
				this.showError = true
				return
			}
			this.showError = false

			this.labelService
				.create(this.label)
				.then((response) => {
					this.$router.push({
						name: 'labels.index',
						params: { id: response.id },
					})
					this.success(
						{ message: 'The label was successfully created.' },
						this
					)
				})
				.catch((e) => {
					this.error(e, this)
				})
		},
	},
}
</script>
