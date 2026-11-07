<template>
	<FormSelect
		v-model.number="percentDone"
		:disabled="disabled"
		:aria-label="$t('task.attributes.percentDone')"
	>
		<option
			v-for="option in PERCENT_OPTIONS"
			:key="option"
			:value="option"
		>
			{{ option * 100 }}%
		</option>
	</FormSelect>
</template>

<script setup lang="ts">
import FormSelect from '@/components/input/FormSelect.vue'

withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const percentDone = defineModel<number>({ required: true })

const PERCENT_OPTIONS = [0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1] as const
</script>
