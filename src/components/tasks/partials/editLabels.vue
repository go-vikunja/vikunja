<template>
	<multiselect
		:loading="loading"
		:placeholder="$t('task.label.placeholder')"
		:multiple="true"
		@search="findLabel"
		:search-results="foundLabels"
		@select="addLabel"
		label="title"
		:creatable="true"
		@create="createAndAddLabel"
		:create-placeholder="$t('task.label.createPlaceholder')"
		v-model="labels"
		:search-delay="10"
	>
		<template #tag="props">
			<span
				:style="{'background': props.item.hexColor, 'color': props.item.textColor}"
				class="tag">
				<span>{{ props.item.title }}</span>
				<a @click="removeLabel(props.item)" class="delete is-small"></a>
			</span>
		</template>
		<template #searchResult="props">
			<span
				v-if="typeof props.option === 'string'"
				class="tag">
				<span>{{ props.option }}</span>
			</span>
			<span
				v-else
				:style="{'background': props.option.hexColor, 'color': props.option.textColor}"
				class="tag">
				<span>{{ props.option.title }}</span>
			</span>
		</template>
	</multiselect>
</template>

<script>
import differenceWith from 'lodash/differenceWith'

import LabelModel from '../../../models/label'
import LabelTaskService from '../../../services/labelTask'

import Multiselect from '@/components/input/multiselect.vue'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'

export default {
	name: 'edit-labels',
	props: {
		value: {
			default: () => [],
			type: Array,
		},
		taskId: {
			type: Number,
			required: false,
			default: () => 0,
		},
		disabled: {
			default: false,
		},
	},
	data() {
		return {
			labelTaskService: LabelTaskService,
			labelTimeout: null,
			labels: [],
			query: '',
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
		this.labelTaskService = new LabelTaskService()
		this.labels = this.value
	},
	computed: {
		foundLabels() {
			const labels = (Object.values(this.$store.state.labels.labels).filter(l => {
				return l.title.toLowerCase().includes(this.query.toLowerCase())
			}) ?? [])

			return differenceWith(labels, this.labels, (first, second) => {
				return first.id === second.id
			})
		},
		loading() {
			return this.labelTaskService.loading || (this.$store.state[LOADING] && this.$store.state[LOADING_MODULE] === 'labels')
		},
	},
	methods: {
		findLabel(query) {
			this.query = query
		},
		addLabel(label, showNotification = true) {
			const bubble = () => {
				this.$emit('input', this.labels)
				this.$emit('change', this.labels)
			}
			
			if (this.taskId === 0) {
				bubble()
				return
			}

			this.$store.dispatch('tasks/addLabel', {label: label, taskId: this.taskId})
				.then(() => {
					bubble()
					if (showNotification) {
						this.$message.success({message: this.$t('task.label.addSuccess')})
					}
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
		removeLabel(label) {
			const removeFromState = () => {
				for (const l in this.labels) {
					if (this.labels[l].id === label.id) {
						this.labels.splice(l, 1)
					}
				}
				this.$emit('input', this.labels)
				this.$emit('change', this.labels)
			}

			if (this.taskId === 0) {
				removeFromState()
				return
			}

			this.$store.dispatch('tasks/removeLabel', {label: label, taskId: this.taskId})
				.then(() => {
					removeFromState()
					this.$message.success({message: this.$t('task.label.removeSuccess')})
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
		createAndAddLabel(title) {
			if (this.taskId === 0) {
				return
			}

			const newLabel = new LabelModel({title: title})
			this.$store.dispatch('labels/createLabel', newLabel)
				.then(r => {
					this.addLabel(r, false)
					this.labels.push(r)
					this.$message.success({message: this.$t('task.label.removeSuccess')})
				})
				.catch(e => {
					this.$message.error(e)
				})
		},

	},
}
</script>
