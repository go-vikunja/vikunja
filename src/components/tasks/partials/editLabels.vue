<template>
	<multiselect
			:multiple="true"
			:close-on-select="false"
			:clear-on-select="true"
			:options-limit="300"
			:hide-selected="true"
			v-model="labels"
			:options="foundLabels"
			:searchable="true"
			:loading="labelService.loading || labelTaskService.loading"
			:internal-search="true"
			@search-change="findLabel"
			@select="addLabel"
			placeholder="Type to add a new label..."
			label="title"
			track-by="id"
			:taggable="true"
			:showNoOptions="false"
			@tag="createAndAddLabel"
			tag-placeholder="Add this as new label"
	>
		<template slot="tag" slot-scope="{ option }">
						<span class="tag"
							:style="{'background': option.hexColor, 'color': option.textColor}">
							<span>{{ option.title }}</span>
							<a class="delete is-small" @click="removeLabel(option)"></a>
						</span>
		</template>
		<template slot="clear" slot-scope="props">
			<div class="multiselect__clear" v-if="labels.length"
				@mousedown.prevent.stop="clearAllLabels(props.search)"></div>
		</template>
	</multiselect>
</template>

<script>
	import { differenceWith } from 'lodash'
	import multiselect from 'vue-multiselect'

	import LabelService from '../../../services/label'
	import LabelModel from '../../../models/label'
	import LabelTaskService from '../../../services/labelTask'

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
			multiselect,
		},
		watch: {
			value(newLabels) {
				this.labels = newLabels
			}
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
			addLabel(label) {
				this.$store.dispatch('tasks/addLabel', {label: label, taskId: this.taskId})
					.then(() => {
						this.$emit('input', this.labels)
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
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			createAndAddLabel(title) {
				let newLabel = new LabelModel({title: title})
				this.labelService.create(newLabel)
					.then(r => {
						this.addLabel(r)
						this.labels.push(r)
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