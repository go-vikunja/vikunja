<template>
	<Multiselect
		class="control is-expanded"
		:placeholder="$t('project.search')"
		:search-results="foundProjects"
		label="title"
		:select-placeholder="$t('project.searchSelect')"
		:model-value="project"
		@update:modelValue="Object.assign(project, $event)"
		@select="select"
		@search="findProjects"
	>
		<template #searchResult="{option}">
			<span
				v-if="projectStore.getAncestors(option).length > 1"
				class="has-text-grey"
			>
				{{ projectStore.getAncestors(option).filter(p => p.id !== option.id).map(p => getProjectTitle(p)).join(' &gt; ') }} &gt;
			</span>
			{{ getProjectTitle(option) }}
		</template>
	</Multiselect>
</template>

<script lang="ts" setup>
import {reactive, ref, watch} from 'vue'

import type {IProject} from '@/modelTypes/IProject'

import {useProjectStore} from '@/stores/projects'
import {getProjectTitle} from '@/helpers/getProjectTitle'

import ProjectModel from '@/models/project'
import Multiselect from '@/components/input/Multiselect.vue'

const props = withDefaults(defineProps<{
	modelValue?: IProject
	savedFiltersOnly?: boolean
}>(), {
	modelValue: new ProjectModel(),
	savedFiltersOnly: false,
})

const emit = defineEmits(['update:modelValue'])

const project = reactive<IProject>(new ProjectModel())

watch(
	() => props.modelValue,
	(newProject) => Object.assign(project, newProject),
	{
		immediate: true,
		deep: true,
	},
)

const projectStore = useProjectStore()
const foundProjects = ref<IProject[]>([])
function findProjects(query: string) {
	if (query === '') {
		select(null)
	}
	
	if (props.savedFiltersOnly) {
		foundProjects.value = projectStore.searchSavedFilter(query)
		return
	}
	
	foundProjects.value = projectStore.searchProject(query)
}

function select(p: IProject | null) {
	if (p === null) {
		Object.assign(project, {id: 0})
	}
	Object.assign(project, p)
	emit('update:modelValue', project)
}
</script>
