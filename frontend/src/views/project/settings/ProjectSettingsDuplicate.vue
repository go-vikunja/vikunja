<template>
	<CreateEdit
		v-model:loading="loadingModel"
		:title="$t('project.duplicate.title')"
		primary-icon="paste"
		:primary-label="$t('project.duplicate.label')"
		:primary-disabled="Boolean(loadError)"
		@primary="duplicate"
	>
		<ErrorMessage v-if="loadError" />
		<template v-else>
			<p>{{ $t('project.duplicate.text') }}</p>
			<ProjectSearch v-model="parentProject" />
			<FancyCheckbox
				v-model="duplicateShares"
				class="mbs-2"
			>
				{{ $t('project.duplicate.shares') }}
			</FancyCheckbox>
		</template>
	</CreateEdit>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import ErrorMessage from '@/components/misc/Error.vue'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProject} from '@/composables/useProject'
import {useProjectNavigation} from '@/composables/useProjectNavigation'
import type {ProjectResponse} from '@/client/queries/projects'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('project.duplicate.title'))

const route = useRoute()
const projectStore = useProjectNavigation()

const {project, isLoading, duplicateProject, error: loadError} = useProject(route.params.projectId)

const parentProject = ref<ProjectResponse | null>(null)
const duplicateShares = ref(true)
const isDuplicating = ref(false)

const loadingModel = computed({
	get: () => isDuplicating.value || isLoading.value,
	set(value: boolean) {
		isDuplicating.value = value
	},
})
watch(
	() => projectStore.projects[project.value.parent_project_id],
	parent => parentProject.value = parent ?? null,
	{immediate: true},
)

async function duplicate() {
	isDuplicating.value = true

	try {
		await duplicateProject(parentProject.value?.id ?? 0, duplicateShares.value)
		success({message: t('project.duplicate.success')})
	} finally {
		isDuplicating.value = false
	}
}
</script>
