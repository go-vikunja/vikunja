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
import type {IProject} from '@/modelTypes/IProject'
import {useSavedFilter} from '@/services/savedFilter'

const props = defineProps<{
	projectId: IProject['id'],
}>()

const {deleteFilter} = useSavedFilter(() => props.projectId)
</script>
