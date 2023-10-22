<template>
	<modal
		@close="$router.back()"
	>
		<card
			:title="project?.title"
		>
			<div class="has-text-left" v-html="htmlDescription" v-if="htmlDescription !== ''"></div>
			<p v-else class="is-italic">
				{{ $t('project.noDescriptionAvailable') }}
			</p>
		</card>
	</modal>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import DOMPurify from 'dompurify'
import {useProjectStore} from '@/stores/projects'

const props = defineProps({
	projectId: {
		type: Number,
		required: true,
	},
})

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
