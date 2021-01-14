<template>
	<multiselect
		:loading="labelService.loading || labelTaskService.loading"
		placeholder="Type to add a new label..."
		:multiple="true"
		@search="findLabel"
		:search-results="foundLabels"
		@select="addLabel"
		label="title"
		:creatable="true"
		@create="createAndAddLabel"
		create-placeholder="Add this as new label"
		v-model="labels"
	>
		<template v-slot:tag="props">
			<span
				:style="{'background': props.item.hexColor, 'color': props.item.textColor}"
				class="tag ml-2 mt-2">
				<span>{{ props.item.title }}</span>
				<a @click="removeLabel(props.item)" class="delete is-small"></a>
			</span>
		</template>
		<template v-slot:searchResult="props">
			<span
				v-if="typeof props.option === 'string'"
				class="tag ml-2">
				<span>{{ props.option }}</span>
			</span>
			<span
				v-else
				:style="{'background': props.option.hexColor, 'color': props.option.textColor}"
				class="tag ml-2">
				<span>{{ props.option.title }}</span>
			</span>
		</template>
	</multiselect>
</template>

<script>
import differenceWith from 'lodash/differenceWith'

import LabelService from '../../../services/label'
import LabelModel from '../../../models/label'
import LabelTaskService from '../../../services/labelTask'

import Multiselect from '@/components/input/multiselect'

export default {
	name: 'edit-labels',
	props: {
		value: {
			default: () => [],
			type: Array,
		},
		taskId: {
			type: Number,
			required: true,
		},
		disabled: {
			default: false,
		},
	},
	data() {
		return {
			labelService: LabelService,
			labelTaskService: LabelTaskService,
			foundLabels: [],
			labelTimeout: null,
			labels: [],
			searchQuery: '',
		}
	},
	components: {
		Multiselect,
	},
	watch: {
		value(newLabels) {
			this.labels = newLabels
		},
	},
	created() {
		this.labelService = new LabelService()
		this.labelTaskService = new LabelTaskService()
		this.labels = this.value
	},
	methods: {
		findLabel(query) {
			this.searchQuery = query
			if (query === '') {
				this.clearAllLabels()
				return
			}

			if (this.labelTimeout !== null) {
				clearTimeout(this.labelTimeout)
			}

			// Delay the search 300ms to not send a request on every keystroke
			this.labelTimeout = setTimeout(() => {
				this.labelService.getAll({}, {s: query})
					.then(response => {
						this.$set(this, 'foundLabels', differenceWith(response, this.labels, (first, second) => {
							return first.id === second.id
						}))
						this.labelTimeout = null
					})
					.catch(e => {
						this.error(e, this)
					})
			}, 300)
		},
		clearAllLabels() {
			this.$set(this, 'foundLabels', [])
		},
		addLabel(label, showNotification = true) {
			this.$store.dispatch('tasks/addLabel', {label: label, taskId: this.taskId})
				.then(() => {
					this.$emit('input', this.labels)
					if (showNotification) {
						this.success({message: 'The label has been added successfully.'}, this)
					}
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		removeLabel(label) {
			this.$store.dispatch('tasks/removeLabel', {label: label, taskId: this.taskId})
				.then(() => {
					// Remove the label from the list
					for (const l in this.labels) {
						if (this.labels[l].id === label.id) {
							this.labels.splice(l, 1)
						}
					}
					this.$emit('input', this.labels)
					this.success({message: 'The label has been removed successfully.'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		createAndAddLabel(title) {
			let newLabel = new LabelModel({title: title})
			this.labelService.create(newLabel)
				.then(r => {
					this.addLabel(r, false)
					this.labels.push(r)
					this.success({message: 'The label has been created successfully.'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},

	},
}
</script>

<style scoped>

</style>