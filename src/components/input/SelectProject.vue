<template>
	<multiselect
		v-model="selectedProjects"
		:search-results="foundProjects"
		:loading="projectService.loading"
		:multiple="true"
		:placeholder="$t('project.search')"
		label="title"
		@search="findProjects"
	/>
</template>

<script setup lang="ts">
import {computed, ref, shallowReactive, watchEffect, type PropType} from 'vue'

import Multiselect from '@/components/input/multiselect.vue'

import type {IProject} from '@/modelTypes/IProject'

import ProjectService from '@/services/project'
import {includesById} from '@/helpers/utils'

const props = defineProps({
	modelValue: {
		type: Array as PropType<IProject[]>,
		default: () => [],
	},
})
const emit = defineEmits<{
	(e: 'update:modelValue', value: IProject[]): void
}>()

const projects = ref<IProject[]>([])

watchEffect(() => {
	projects.value = props.modelValue
})

const selectedProjects = computed({
	get() {
		return projects.value
	},
	set: (value) => {
		projects.value = value
		emit('update:modelValue', value)
	},
})

const projectService = shallowReactive(new ProjectService())
const foundProjects = ref<IProject[]>([])

async function findProjects(query: string) {
	if (query === '') {
		foundProjects.value = []
		return
	}

	const response = await projectService.getAll({}, {s: query}) as IProject[]

	// Filter selected items from the results
	foundProjects.value = response.filter(({id}) => !includesById(projects.value, id))
}
</script>