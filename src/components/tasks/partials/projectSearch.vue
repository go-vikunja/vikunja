<template>
	<Multiselect
		class="control is-expanded"
		:placeholder="$t('project.search')"
		:search-results="foundProjects"
		label="title"
		:select-placeholder="$t('project.searchSelect')"
		:model-value="project"
		@update:model-value="Object.assign(project, $event)"
		@select="select"
		@search="findProjects"
	>
		<template #searchResult="{option}">
			{{ (option as IProject).title }}
		</template>
	</Multiselect>
</template>

<script lang="ts" setup>
import {reactive, ref, watch} from 'vue'
import type {PropType} from 'vue'
import {useI18n} from 'vue-i18n'

import type {IProject} from '@/modelTypes/IProject'

import {useProjectStore} from '@/stores/projects'

import ProjectModel from '@/models/project'

import Multiselect from '@/components/input/multiselect.vue'

const props = defineProps({
	modelValue: {
		type: Object as PropType<IProject>,
		required: false,
	},
})
const emit = defineEmits(['update:modelValue'])

const {t} = useI18n({useScope: 'global'})

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
	foundProjects.value = projectStore.searchProject(query)
}

function select(l: IProject | null) {
	if (l === null) {
		return
	}
	Object.assign(project, l)
	emit('update:modelValue', project)
}
</script>
