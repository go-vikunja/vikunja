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
import {useSavedFilter} from '@/services/savedFilter'

const props = defineProps<{
	projectId: number,
}>()

const {deleteFilter} = useSavedFilter(() => props.projectId)
</script>
