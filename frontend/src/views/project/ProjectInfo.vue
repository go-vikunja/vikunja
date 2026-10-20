<!-- eslint-disable vue/no-v-html -->
<template>
	<Modal
		@close="goBack()"
	>
		<Card
			:title="project?.title"
			class="is-justify-content-start"
			:show-close="true"
			@close="goBack()"
		>
			<div
				v-if="htmlDescription !== ''"
				class="has-text-start"
				data-user-content
				v-html="htmlDescription"
			/>
			<p
				v-else
				class="is-italic"
			>
				{{ $t('project.noDescriptionAvailable') }}
			</p>
		</Card>
	</Modal>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import {useProjects} from '@/composables/useProjects'
import {useGoBack} from '@/composables/useGoBack'
import {useI18n} from 'vue-i18n'
import {sanitizeProjectDescription} from '@/helpers/sanitizeProjectDescription'

const props = defineProps<{
	projectId: number
}>()

const {t} = useI18n()
const goBack = useGoBack()

const projectList = useProjects()
const project = computed(() => projectList.projects[props.projectId])
const htmlDescription = computed(() => {
	const description = project.value?.description || ''
	if (description === '') {
		return ''
	}

	if (project.value.id === -1) {
		return t('project.favoriteDescription')
	}

	return sanitizeProjectDescription(description)
})
</script>
