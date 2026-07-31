<!-- eslint-disable vue/no-v-html -->
<template>
	<Modal
		@close="$router.back()"
	>
		<Card
			:title="project?.title"
			class="is-justify-content-start"
			:show-close="true"
			@close="$router.back()"
		>
			<div
				v-if="htmlDescription !== ''"
				class="has-text-start"
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
import {useI18n} from 'vue-i18n'

const props = defineProps<{
	projectId: number
}>()

const {t} = useI18n()

// ADD_ATTR re-permits target, which lets a link hand the opened page a live
// window.opener back to this tab. rel must be forced here rather than allowed,
// so a description carrying its own rel can't drop noopener.
DOMPurify.addHook('afterSanitizeAttributes', node => {
	if (node.hasAttribute('target')) {
		node.setAttribute('rel', 'noopener noreferrer')
	}
})

const projectStore = useProjectStore()
const project = computed(() => projectStore.projects[props.projectId])
const htmlDescription = computed(() => {
	const description = project.value?.description || ''
	if (description === '') {
		return ''
	}

	if (project.value.id === -1) {
		return t('project.favoriteDescription')
	}

	return DOMPurify.sanitize(description, {ADD_ATTR: ['target']})
})
</script>
