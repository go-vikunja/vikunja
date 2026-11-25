<template>
	<Multiselect
		class="control is-expanded"
		:placeholder="$t('project.search')"
		:search-results="foundProjects"
		label="title"
		:select-placeholder="$t('project.searchSelect')"
		:model-value="multiple ? modelValue as ProjectResponse[] : project"
		:multiple="multiple"
		:close-after-select="!multiple"
		@update:modelValue="update"
		@select="select"
		@search="findProjects"
	>
		<template #searchResult="{option}">
			<span
				v-if="projectList.getAncestors(option).length > 1"
				class="has-text-grey"
			>
				{{ projectList.getAncestors(option).slice(0, -1).map(p => getProjectTitle(p)).join(' &gt; ') }} &gt;
			</span>
			{{ getProjectTitle(option) }}
		</template>
	</Multiselect>
</template>

<script lang="ts" setup generic="Multiple extends boolean = false">
import {reactive, ref, watch} from 'vue'

import {normalizeProject, type ProjectResponse} from '@/client/queries/projects'
import {useProjects} from '@/composables/useProjects'
import {getProjectTitle} from '@/helpers/getProjectTitle'

import Multiselect from '@/components/input/Multiselect.vue'

type Selection = Multiple extends true ? ProjectResponse[] : ProjectResponse | null

const props = withDefaults(defineProps<{
	modelValue?: Selection
	// `& boolean` keeps Vue's runtime Boolean casting for a bare `multiple` attribute.
	multiple?: Multiple & boolean
	savedFiltersOnly?: boolean
	filter?: (project: ProjectResponse) => boolean,
}>(), {
	modelValue: undefined,
	multiple: undefined,
	savedFiltersOnly: false,
	filter: () => true,
})

const emit = defineEmits<{
	'update:modelValue': [value: Selection]
}>()

const project = reactive<ProjectResponse>(normalizeProject({id: 0}))

watch(
	() => props.multiple ? null : props.modelValue as ProjectResponse | null | undefined,
	(newProject) => Object.assign(project, newProject ?? normalizeProject({id: 0})),
	{
		immediate: true,
		deep: true,
	},
)

const projectList = useProjects()

const foundProjects = ref<ProjectResponse[]>([])
function findProjects(query: string) {
	if (query === '' && !props.multiple) {
		select(null)
	}
	
	if (props.savedFiltersOnly) {
		const found = projectList.searchSavedFilter(query)
		foundProjects.value = found.filter(props.filter)
		return
	}
	
	const found = projectList.searchProject(query)
	foundProjects.value = found.filter(props.filter)
}

function update(value: ProjectResponse | ProjectResponse[] | null) {
	if (Array.isArray(value)) {
		emit('update:modelValue', value as Selection)
	} else if (value === null) {
		select(null)
	} else {
		Object.assign(project, value)
	}
}

function select(p: ProjectResponse | null) {
	if (props.multiple) {
		return
	}
	if (p === null) {
		Object.assign(project, normalizeProject({id: 0}))
		emit('update:modelValue', null as Selection)
		return
	}
	Object.assign(project, p)
	emit('update:modelValue', project as Selection)
}
</script>
