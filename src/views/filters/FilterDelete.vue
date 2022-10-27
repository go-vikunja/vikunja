<template>
	<modal
		@close="$router.back()"
		@submit="deleteFilter()"
	>
		<template #header>
			<span>{{ $t('filters.delete.header') }}</span>
		</template>
		
		<template #text>
			<p>{{ $t('filters.delete.text') }}</p>
		</template>
	</modal>
</template>

<script setup lang="ts">
import {toRef} from 'vue'
import type {IList} from '@/modelTypes/IList'
import {useSavedFilter} from '@/services/savedFilter'

const props = defineProps<{ listId: IList['id'] }>()

const {deleteFilter} = useSavedFilter(toRef(props, 'listId'))
</script>
