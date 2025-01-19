<template>
	<Modal
		@close="$router.back()"
		@submit="deleteFilter()"
	>
		<template #header>
			<span>{{ $t('filters.delete.header') }}</span>
		</template>

		<template #text>
			<p>{{ $t('filters.delete.text') }}</p>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import type { RouteLocationNormalizedLoaded } from 'vue-router'

import type {IProject} from '@/modelTypes/IProject'
import {useSavedFilter} from '@/services/savedFilter'

definePage({
	name: 'filter.settings.delete',
	meta: { showAsModal: true },
	props: route => {
		// https://github.com/posva/unplugin-vue-router/discussions/513#discussioncomment-10695660
		const castedRoute = route as RouteLocationNormalizedLoaded<'filter.settings.delete'>
		return { projectId: Number(castedRoute.params.projectId) }
	},
})

const props = defineProps<{
	projectId: IProject['id'],
}>()

const {deleteFilter} = useSavedFilter(() => props.projectId)
</script>
