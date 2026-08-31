<template>
	<Multiselect
		class="control is-expanded"
		:placeholder="$t('project.search')"
		:search-results="foundProjects"
		label="title"
		:select-placeholder="$t('project.searchSelect')"
		:model-value="project"
		@update:modelValue="(val) => val === null ? select(null) : Object.assign(project, val)"
		@select="select"
		@search="findProjects"
	>
		<template #searchResult="{option}">
			<span
				v-if="projectAncestors(option).length > 1"
				class="has-text-grey"
			>
				{{ projectAncestors(option).slice(0, -1).map(p => getProjectTitle(p)).join(' &gt; ') }} &gt;
			</span>
			{{ projectOptionTitle(option) }}
		</template>
	</Multiselect>
</template>

<script lang="ts" setup>
import {reactive, ref, watch} from 'vue'

import {normalizeProject, type ProjectResponse} from '@/client/queries/projects'
import {useProjectNavigation} from '@/composables/useProjectNavigation'
import {getProjectTitle} from '@/helpers/getProjectTitle'

import Multiselect from '@/components/input/Multiselect.vue'

const props = withDefaults(defineProps<{
	modelValue?: ProjectResponse | null
	savedFiltersOnly?: boolean
	filter?: (project: ProjectResponse) => boolean,
}>(), {
	modelValue: () => normalizeProject({id: 0}),
	savedFiltersOnly: false,
	filter: () => true,
})

const emit = defineEmits<{
	'update:modelValue': [value: ProjectResponse | null]
}>()

const project = reactive<ProjectResponse>(normalizeProject({id: 0}))

watch(
	() => props.modelValue,
	(newProject) => Object.assign(project, newProject ?? normalizeProject({id: 0})),
	{
		immediate: true,
		deep: true,
	},
)

const projectNavigation = useProjectNavigation()

function projectAncestors(option: unknown): ProjectResponse[] {
	if (typeof option !== 'object' || option === null || !('id' in option)) {
		return []
	}

	return projectNavigation.getAncestors(option as ProjectResponse)
}

function projectOptionTitle(option: unknown): string {
	if (typeof option !== 'object' || option === null || !('id' in option) || !('title' in option)) {
		return String(option ?? '')
	}

	return getProjectTitle(option as ProjectResponse)
}

const foundProjects = ref<ProjectResponse[]>([])
function findProjects(query: string) {
	if (query === '') {
		select(null)
	}
	
	if (props.savedFiltersOnly) {
		const found = projectNavigation.searchSavedFilter(query)
		foundProjects.value = found.filter(props.filter)
		return
	}
	
	const found = projectNavigation.searchProject(query)
	foundProjects.value = found.filter(props.filter)
}

function select(p: ProjectResponse | null) {
	if (p === null) {
		Object.assign(project, normalizeProject({id: 0}))
		emit('update:modelValue', null)
		return
	}
	Object.assign(project, p)
	emit('update:modelValue', project)
}
</script>
