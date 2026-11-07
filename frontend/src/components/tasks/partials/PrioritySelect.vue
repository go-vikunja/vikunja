<template>
	<FormSelect
		v-model.number="priority"
		:disabled="disabled"
		flat
		:aria-label="$t('task.attributes.priority')"
	>
		<option :value="PRIORITIES.UNSET">
			{{ $t('task.priority.unset') }}
		</option>
		<option :value="PRIORITIES.LOW">
			{{ $t('task.priority.low') }}
		</option>
		<option :value="PRIORITIES.MEDIUM">
			{{ $t('task.priority.medium') }}
		</option>
		<option :value="PRIORITIES.HIGH">
			{{ $t('task.priority.high') }}
		</option>
		<option :value="PRIORITIES.URGENT">
			{{ $t('task.priority.urgent') }}
		</option>
		<option :value="PRIORITIES.DO_NOW">
			{{ $t('task.priority.doNow') }}
		</option>
	</FormSelect>
</template>

<script setup lang="ts">
import FormSelect from '@/components/input/FormSelect.vue'
import {PRIORITIES} from '@/constants/priorities'

withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const priority = defineModel<number>({
	required: true,
	default: 0,
})

</script>
