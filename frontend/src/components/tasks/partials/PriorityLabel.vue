<template>
	<span
		v-if="!done && (showAll || priority >= minimumPriority)"
		:class="{
			'negligible': priority <= priorities.LOW,
			'not-so-high': priority > priorities.LOW && priority < priorities.HIGH,
			'high-priority': priority >= priorities.HIGH
		}"
		class="priority-label"
	>
		<span class="icon">
			<Icon
				v-if="priority >= priorities.HIGH"
				icon="exclamation-circle"
			/>
			<Icon
				v-else
				icon="exclamation"
			/>
		</span>
		<span>
			<template v-if="priority === priorities.UNSET">{{ $t('task.priority.unset') }}</template>
			<template v-if="priority === priorities.LOW">{{ $t('task.priority.low') }}</template>
			<template v-if="priority === priorities.MEDIUM">{{ $t('task.priority.medium') }}</template>
			<template v-if="priority === priorities.HIGH">{{ $t('task.priority.high') }}</template>
			<template v-if="priority === priorities.URGENT">{{ $t('task.priority.urgent') }}</template>
			<template v-if="priority === priorities.DO_NOW">{{ $t('task.priority.doNow') }}</template>
		</span>
	</span>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {PRIORITIES as priorities} from '@/constants/priorities'
import {useAuthStore} from '@/stores/auth'
	
withDefaults(defineProps<{
	priority: number,
	showAll?: boolean,
	done?: boolean
}>(), {
	showAll: false,
	done: false,
})

const authStore = useAuthStore()

const minimumPriority = computed(() => {
	return authStore.settings.frontendSettings.minimumPriority
})
</script>

<style lang="scss" scoped>
.high-priority {
	color: var(--danger);
	inline-size: auto !important; // To override the width set in tasks
}

.not-so-high {
	color: var(--warning);
}

.negligible {
	color: var(--info);
}

.icon {
	vertical-align: top;
	inline-size: auto !important;
	padding-inline-end: .5rem;
}
</style>
