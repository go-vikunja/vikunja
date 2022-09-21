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
				<BaseButton v-cy="'taskDetail.removeLabel'" @click="removeLabel(label)" class="delete is-small" />
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
import {type PropType, ref, computed, shallowReactive, watch} from 'vue'
import {useStore} from '@/store'
import {useI18n} from 'vue-i18n'

import LabelModel from '@/models/label'
import LabelTaskService from '@/services/labelTask'
import {success} from '@/message'

import BaseButton from '@/components/base/BaseButton.vue'
import Multiselect from '@/components/input/multiselect.vue'
import type { ILabel } from '@/modelTypes/ILabel'
import { useLabelStore } from '@/stores/labels'

const props = defineProps({
	modelValue: {
		type: Array as PropType<ILabel[]>,
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

const emit = defineEmits(['update:modelValue'])

const store = useStore()
const {t} = useI18n({useScope: 'global'})

const labelTaskService = shallowReactive(new LabelTaskService())
const labels = ref<ILabel[]>([])
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

const labelStore = useLabelStore()

const foundLabels = computed(() => labelStore.filterLabelsByQuery(labels.value, query.value))
const loading = computed(() => labelTaskService.loading || labelStore.isLoading)

function findLabel(newQuery: string) {
	query.value = newQuery
}

async function addLabel(label: ILabel, showNotification = true) {
	const bubble = () => {
		emit('update:modelValue', labels.value)
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

async function removeLabel(label: ILabel) {
	if (props.taskId !== 0) {
		await store.dispatch('tasks/removeLabel', {label, taskId: props.taskId})
	}

	for (const l in labels.value) {
		if (labels.value[l].id === label.id) {
			labels.value.splice(l, 1) // FIXME: l should be index
		}
	}
	emit('update:modelValue', labels.value)
	success({message: t('task.label.removeSuccess')})
}

async function createAndAddLabel(title: string) {
	if (props.taskId === 0) {
		return
	}

	const labelStore = useLabelStore()
	const newLabel = await labelStore.createLabel(new LabelModel({title}))
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
