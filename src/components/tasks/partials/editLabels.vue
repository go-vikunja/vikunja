<template>
	<multiselect
		:clear-on-select="true"
		:close-on-select="false"
		:disabled="disabled"
		:hide-selected="true"
		:internal-search="true"
		:loading="labelService.loading || labelTaskService.loading"
		:multiple="true"
		:options="foundLabels"
		:options-limit="300"
		:searchable="true"
		:showNoOptions="false"
		:taggable="true"
		@search-change="findLabel"
		@select="label => addLabel(label)"
		@tag="createAndAddLabel"
		label="title"
		placeholder="Type to add a new label..."
		tag-placeholder="Add this as new label"
		track-by="id"
		v-model="labels"
	>
		<template
			slot="tag"
			slot-scope="{ option }">
			<span
				:style="{'background': option.hexColor, 'color': option.textColor}"
				class="tag">
				<span>{{ option.title }}</span>
				<a @click="removeLabel(option)" class="delete is-small"></a>
			</span>
		</template>
		<template slot="clear" slot-scope="props">
			<div
				@mousedown.prevent.stop="clearAllLabels(props.search)"
				class="multiselect__clear"
				v-if="labels.length"></div>
		</template>
	</multiselect>
</template>

<script>
import differenceWith from 'lodash/differenceWith'

import LabelService from '../../../services/label'
import LabelModel from '../../../models/label'
import LabelTaskService from '../../../services/labelTask'
import LoadingComponent from '../../misc/loading'
import ErrorComponent from '../../misc/error'

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
		multiselect: () => ({
			component: import(/* webpackChunkName: "multiselect" */ 'vue-multiselect'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
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