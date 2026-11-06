<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import Multiselect from '@/components/input/Multiselect.vue'
import {useProjects} from '@/composables/useProjects'

const ids = defineModel<number[]>({required: true})
const {t} = useI18n({useScope: 'global'})
const projects = useProjects()
const query = ref('')
const selected = computed(() => ids.value.map(id => ({
	id,
	title: projects.projects[id]?.title ?? t('filters.simple.unknownProject', {id}),
})))
const results = computed(() => projects.projectsArray
	.filter(project => project.id > 0 && !project.is_archived && !ids.value.includes(project.id) &&
	project.title.toLowerCase().includes(query.value.toLowerCase()))
	.map(project => ({
		id: project.id,
		title: project.title,
	})))
</script>

<template>
	<Multiselect
		:model-value="selected"
		:search-results="results"
		:loading="projects.isLoading"
		:multiple="true"
		:show-empty="true"
		:close-after-select="false"
		:aria-label="t('project.projects')"
		:placeholder="t('project.search')"
		label="title"
		@search="query = $event"
		@update:modelValue="ids = Array.isArray($event) ? $event.map(project => project.id) : []"
	/>
</template>
