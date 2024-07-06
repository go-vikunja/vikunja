<template>
	<div class="select">
		<select
			v-model.number="percentDone"
			:disabled="disabled || undefined"
		>
			<option
				v-for="option in PERCENT_OPTIONS"
				:key="option"
				:value="option"
			>
				{{ option * 100 }}%
			</option>
		</select>
	</div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const percentDone = defineModel<number>({ required: true })

const PERCENT_OPTIONS = [0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1] as const
</script>
