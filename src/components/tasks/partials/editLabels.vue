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
		:close-after-select="false"
	>
		<template #tag="props">
			<span
				:style="{'background': props.item.hexColor, 'color': props.item.textColor}"
				class="tag">
				<span>{{ props.item.title }}</span>
				<button type="button" v-cy="'taskDetail.removeLabel'" @click="removeLabel(props.item)" class="delete is-small" />
			</span>
		</template>
		<template #searchResult="props">
			<span
				v-if="typeof props.option === 'string'"
				class="tag search-result">
				<span>{{ props.option }}</span>
			</span>
			<span
				v-else
				:style="{'background': props.option.hexColor, 'color': props.option.textColor}"
				class="tag search-result">
				<span>{{ props.option.title }}</span>
			</span>
		</template>
	</multiselect>
</template>

<script lang="ts">
import LabelModel from '../../../models/label'
import LabelTaskService from '../../../services/labelTask'

import Multiselect from '@/components/input/multiselect.vue'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'

export default {
	name: 'edit-labels',
	props: {
		modelValue: {
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
	emits: ['update:modelValue', 'change'],
	data() {
		return {
			labelTaskService: new LabelTaskService(),
			labelTimeout: null,
			labels: [],
			query: '',
		}
	},
	components: {
		Multiselect,
	},
	watch: {
		modelValue: {
			handler(value) {
				this.labels = value
			},
			immediate: true,
			deep: true,
		},
	},
	computed: {
		foundLabels() {
			return this.$store.getters['labels/filterLabelsByQuery'](this.labels, this.query)
		},
		loading() {
			return this.labelTaskService.loading || (this.$store.state[LOADING] && this.$store.state[LOADING_MODULE] === 'labels')
		},
	},
	methods: {
		findLabel(query) {
			this.query = query
		},

		async addLabel(label, showNotification = true) {
			const bubble = () => {
				this.$emit('update:modelValue', this.labels)
				this.$emit('change', this.labels)
			}
			
			if (this.taskId === 0) {
				bubble()
				return
			}

			await this.$store.dispatch('tasks/addLabel', {label: label, taskId: this.taskId})
			bubble()
			if (showNotification) {
				this.$message.success({message: this.$t('task.label.addSuccess')})
			}
		},

		async removeLabel(label) {
			if (this.taskId !== 0) {
				await this.$store.dispatch('tasks/removeLabel', {label: label, taskId: this.taskId})
			}

			for (const l in this.labels) {
				if (this.labels[l].id === label.id) {
					this.labels.splice(l, 1)
				}
			}
			this.$emit('update:modelValue', this.labels)
			this.$emit('change', this.labels)
			this.$message.success({message: this.$t('task.label.removeSuccess')})
		},

		async createAndAddLabel(title) {
			if (this.taskId === 0) {
				return
			}

			const newLabel = new LabelModel({title: title})
			const label = await this.$store.dispatch('labels/createLabel', newLabel)
			this.addLabel(label, false)
			this.labels.push(label)
			this.$message.success({message: this.$t('task.label.addCreateSuccess')})
		},

	},
}
</script>

<style lang="scss" scoped>
.tag {
	margin: .25rem !important;
}

.tag.search-result {
	margin: 0 !important;
}

:deep(.input-wrapper) {
	padding: .25rem !important;
}

:deep(input.input) {
	padding: 0 .5rem;
}
</style>
