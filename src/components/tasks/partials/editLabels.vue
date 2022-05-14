<template>
	<Multiselect
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
		<template #tag="{item: label}">
			<span
				:style="{'background': label.hexColor, 'color': label.textColor}"
				class="tag">
				<span>{{ label.title }}</span>
				<button type="button" v-cy="'taskDetail.removeLabel'" @click="removeLabel(label)" class="delete is-small" />
			</span>
		</template>
		<template #searchResult="{option}">
			<span
				v-if="typeof option === 'string'"
				class="tag search-result">
				<span>{{ option }}</span>
			</span>
			<span
				v-else
				:style="{'background': option.hexColor, 'color': option.textColor}"
				class="tag search-result">
				<span>{{ option.title }}</span>
			</span>
		</template>
	</Multiselect>
</template>

<script setup lang="ts">
import {PropType, ref, computed, shallowReactive, watch} from 'vue'
import {useStore} from 'vuex'
import {useI18n} from 'vue-i18n'

import LabelModel from '@/models/label'
import LabelTaskService from '@/services/labelTask'
import {success} from '@/message'

import Multiselect from '@/components/input/multiselect.vue'

const props = defineProps({
	modelValue: {
		type: Array as PropType<LabelModel[]>,
		default: () => [],
	},
	taskId: {
		type: Number,
		required: false,
		default: 0,
	},
	disabled: {
		default: false,
	},
})

const emit = defineEmits(['update:modelValue', 'change'])

const store = useStore()
const {t} = useI18n()

const labelTaskService = shallowReactive(new LabelTaskService())
const labels = ref<LabelModel[]>([])
const query = ref('')

watch(
	() => props.modelValue,
	(value) => {
		labels.value = value
	},
	{
		immediate: true,
		deep: true,
	},
)

const foundLabels = computed(() => store.getters['labels/filterLabelsByQuery'](labels.value, query.value))
const loading = computed(() => labelTaskService.loading || (store.state.loading && store.state.loadingModule === 'labels'))

function findLabel(newQuery: string) {
	query.value = newQuery
}

async function addLabel(label: LabelModel, showNotification = true) {
	const bubble = () => {
		emit('update:modelValue', labels.value)
		emit('change', labels.value)
	}
	
	if (props.taskId === 0) {
		bubble()
		return
	}

	await store.dispatch('tasks/addLabel', {label, taskId: props.taskId})
	bubble()
	if (showNotification) {
		success({message: t('task.label.addSuccess')})
	}
}

async function removeLabel(label: LabelModel) {
	if (props.taskId !== 0) {
		await store.dispatch('tasks/removeLabel', {label, taskId: props.taskId})
	}

	for (const l in labels.value) {
		if (labels.value[l].id === label.id) {
			labels.value.splice(l, 1)
		}
	}
	emit('update:modelValue', labels.value)
	emit('change', labels.value)
	success({message: t('task.label.removeSuccess')})
}

async function createAndAddLabel(title: string) {
	if (props.taskId === 0) {
		return
	}

	const newLabel = await store.dispatch('labels/createLabel', new LabelModel({title}))
	addLabel(newLabel, false)
	labels.value.push(newLabel)
	success({message: t('task.label.addCreateSuccess')})
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
