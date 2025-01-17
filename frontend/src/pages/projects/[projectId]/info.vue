<template>
	<Modal
		@close="$router.back()"
	>
		<Card
			:title="project?.title"
		>
			<div
				v-if="htmlDescription !== ''"
				class="has-text-left"
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
import DOMPurify from 'dompurify'
import {useProjectStore} from '@/stores/projects'

definePage({
	name: 'project.info',
	meta: { showAsModal: true },
	props: route => ({ projectId: Number(route.params.projectId as string) }),
})

const props = defineProps<{
	projectId: number
}>()

const projectStore = useProjectStore()
const project = computed(() => projectStore.projects[props.projectId])
const htmlDescription = computed(() => {
	const description = project.value?.description || ''
	if (description === '') {
		return ''
	}

	return DOMPurify.sanitize(description, {ADD_ATTR: ['target']})
})
</script>
