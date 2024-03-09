<script setup lang="ts">
import FilterInput from '@/components/project/partials/FilterInput.vue'

function initState(value: string) {
	return {
		value,
	}
}
</script>

<template>
	<Story title="Filter Input">
		<Variant 
			title="With date values"
			:init-state="initState('dueDate < now && done = false && dueDate > now/w+1w')"
		>
			<template #default="{state}">
				<FilterInput v-model="state.value" />
			</template>
		</Variant>
	</Story>
</template>
