<template>
	<Multiselect
		class="control is-expanded"
		:placeholder="$t('project.search')"
		:search-results="foundProjects as unknown as Record<string, unknown>[]"
		label="title"
		:select-placeholder="$t('project.searchSelect')"
		:model-value="project as unknown as Record<string, unknown>"
		@update:modelValue="(value: Record<string, unknown> | Record<string, unknown>[] | null) => value && Object.assign(project, value as unknown as IProject)"
		@select="(value: Record<string, unknown>) => select(value as unknown as IProject)"
		@search="findProjects"
	>
		<template #searchResult="{option}">
			<span
				v-if="projectStore.getAncestors(option as unknown as IProject).length > 1"
				class="has-text-grey"
			>
				{{ projectStore.getAncestors(option as unknown as IProject).filter(p => p.id !== (option as unknown as IProject).id).map(p => getProjectTitle(p)).join(' &gt; ') }} &gt;
			</span>
			{{ getProjectTitle(option as unknown as IProject) }}
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
	filter?: (project: IProject) => boolean,
}>(), {
	modelValue: () => new ProjectModel(),
	savedFiltersOnly: false,
	filter: () => true,
})

const emit = defineEmits<{
	'update:modelValue': [value: IProject]
}>()

const project: IProject = reactive(new ProjectModel())

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
		const found = projectStore.searchSavedFilter(query)
		foundProjects.value = found.filter((p): p is IProject => p !== undefined && props.filter(p))
		return
	}
	
	const found = projectStore.searchProject(query)
	foundProjects.value = found.filter((p): p is IProject => p !== undefined && props.filter(p))
}

function select(p: IProject | null) {
	if (p === null) {
		Object.assign(project, {id: 0})
		return
	}
	Object.assign(project, p)
	emit('update:modelValue', project)
}
</script>
