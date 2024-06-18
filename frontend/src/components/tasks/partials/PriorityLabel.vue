<template>
	<span
		v-if="!done && (showAll || priority >= priorities.HIGH)"
		:class="{'not-so-high': priority === priorities.HIGH, 'high-priority': priority >= priorities.HIGH}"
		class="priority-label"
	>
		<span
			v-if="priority >= priorities.HIGH"
			class="icon"
		>
			<Icon icon="exclamation" />
		</span>
		<span>
			<template v-if="priority === priorities.UNSET">{{ $t('task.priority.unset') }}</template>
			<template v-if="priority === priorities.LOW">{{ $t('task.priority.low') }}</template>
			<template v-if="priority === priorities.MEDIUM">{{ $t('task.priority.medium') }}</template>
			<template v-if="priority === priorities.HIGH">{{ $t('task.priority.high') }}</template>
			<template v-if="priority === priorities.URGENT">{{ $t('task.priority.urgent') }}</template>
			<template v-if="priority === priorities.DO_NOW">{{ $t('task.priority.doNow') }}</template>
		</span>
		<span
			v-if="priority === priorities.DO_NOW"
			class="icon pr-0"
		>
			<Icon icon="exclamation" />
		</span>
	</span>
</template>

<script setup lang="ts">
import {PRIORITIES as priorities} from '@/constants/priorities'
	
defineProps({
	priority: {
		default: 0,
		type: Number,
	},
	showAll: {
		type: Boolean,
		default: false,
	},
	done: {
		type: Boolean,
		default: false,
	},
})
</script>

<style lang="scss" scoped>
span.high-priority {
	color: var(--danger);
	width: auto !important; // To override the width set in tasks

	.icon {
		vertical-align: top;
		width: auto !important;
		padding: 0 .5rem;
	}

	&.not-so-high {
		color: var(--warning);
	}
}
</style>