<template>
	<span
		:class="{'not-so-high': priority === priorities.HIGH, 'high-priority': priority >= priorities.HIGH}"
		class="priority-label"
		v-if="!done && (showAll || priority >= priorities.HIGH)">
		<span class="icon" v-if="priority >= priorities.HIGH">
			<icon icon="exclamation"/>
		</span>
		<span>
			<template v-if="priority === priorities.UNSET">{{ $t('task.priority.unset') }}</template>
			<template v-if="priority === priorities.LOW">{{ $t('task.priority.low') }}</template>
			<template v-if="priority === priorities.MEDIUM">{{ $t('task.priority.medium') }}</template>
			<template v-if="priority === priorities.HIGH">{{ $t('task.priority.high') }}</template>
			<template v-if="priority === priorities.URGENT">{{ $t('task.priority.urgent') }}</template>
			<template v-if="priority === priorities.DO_NOW">{{ $t('task.priority.doNow') }}</template>
		</span>
		<span class="icon" v-if="priority === priorities.DO_NOW">
			<icon icon="exclamation"/>
		</span>
	</span>
</template>

<script lang="ts">
import priorites from '../../../models/constants/priorities'

export default {
	name: 'priorityLabel',
	data() {
		return {
			priorities: priorites,
		}
	},
	props: {
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
	},
}
</script>

<style lang="scss" scoped>
.priority-label {
	display: inline-flex;
	align-items: center;
}

span.high-priority {
	color: var(--danger);
	width: auto !important; // To override the width set in tasks

	.icon {
		vertical-align: middle;
		width: auto !important;
		padding: 0 .5rem;
	}

	&.not-so-high {
		color: var(--warning);
	}
}
</style>