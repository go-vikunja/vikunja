<template>
	<div class="select">
		<select
			v-model="priority"
			class="item"
			:class="mouseOver ? 'item-over' : 'item-out'"
			:disabled="disabled || undefined"
			@mouseover="mouseAction(true)"
			@mouseout="mouseAction(false)"
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
		</select>
	</div>
</template>

<script setup lang="ts">
import {PRIORITIES} from '@/constants/priorities'
import {ref} from "vue";

const mouseOver = ref(false)

const mouseAction = (over: boolean) => {
	mouseOver.value = over
}

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

<style lang="scss" scoped>
.select {
	height: 48px;
	.item {
		background-color: transparent;
		border: none;
		height: 48px;
	}
	.item-out {
		background-color: transparent !important;
	}
	.item-over {
		background-color: transparent
	}
}
</style>
