<template>
	<span
		:class="{'not-so-high': priority === priorities.HIGH, 'high-priority': priority >= priorities.HIGH}"
		class="priority-label"
		v-if="showAll || priority >= priorities.HIGH">
		<span class="icon" v-if="priority >= priorities.HIGH">
			<icon icon="exclamation"/>
		</span>
		<span>
			<template v-if="priority === priorities.UNSET">Unset</template>
			<template v-if="priority === priorities.LOW">Low</template>
			<template v-if="priority === priorities.MEDIUM">Medium</template>
			<template v-if="priority === priorities.HIGH">High</template>
			<template v-if="priority === priorities.URGENT">Urgent</template>
			<template v-if="priority === priorities.DO_NOW">DO NOW</template>
		</span>
		<span class="icon" v-if="priority === priorities.DO_NOW">
			<icon icon="exclamation"/>
		</span>
	</span>
</template>

<script>
import priorites from '../../../models/priorities'

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
	},
}
</script>

<style lang="scss" scoped>
@import '../../../styles/theme/variables/all';

.priority-label {
	display: inline-flex;
	align-items: center;
}

span.high-priority {
	color: $red;
	width: auto !important; // To override the width set in tasks

	.icon {
		vertical-align: middle;
		width: auto !important;
		padding: 0 .5rem;
	}

	&.not-so-high {
		color: $orange;
	}
}
</style>