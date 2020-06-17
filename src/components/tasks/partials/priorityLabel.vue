<template>
	<span v-if="showAll || priority >= priorities.HIGH" :class="{'not-so-high': priority === priorities.HIGH, 'high-priority': priority >= priorities.HIGH}">
		<span class="icon" v-if="priority >= priorities.HIGH">
			<icon icon="exclamation"/>
		</span>
		<template v-if="priority === priorities.UNSET">Unset</template>
		<template v-if="priority === priorities.LOW">Low</template>
		<template v-if="priority === priorities.MEDIUM">Medium</template>
		<template v-if="priority === priorities.HIGH">High</template>
		<template v-if="priority === priorities.URGENT">Urgent</template>
		<template v-if="priority === priorities.DO_NOW">DO NOW</template>
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
		}
	}
</script>

<style lang="scss">
	@import '../../../styles/theme/variables';

	span.high-priority{
		color: $red;
		width: auto !important; // To override the width set in tasks

		.icon {
			vertical-align: middle;
			width: auto !important;
			padding: 0 .5em;
		}

		&.not-so-high {
			color: $orange;
		}
	}
</style>