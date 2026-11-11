<template>
	<XButton
		v-if="type === 'button'"
		v-tooltip="tooltipText"
		variant="secondary"
		:icon="iconName"
		@click="emit('toggle', modelValue === null)"
	>
		{{ buttonText }}
	</XButton>
	<DropdownItem
		v-else-if="type === 'dropdown'"
		v-tooltip="tooltipText"
		:icon="iconName"
		@click="emit('toggle', modelValue === null)"
	>
		{{ buttonText }}
	</DropdownItem>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import DropdownItem from '@/components/misc/DropdownItem.vue'

import type {Subscription} from '@/client/generated'

import type { IconProp } from '@fortawesome/fontawesome-svg-core'

const props = withDefaults(defineProps<{
	modelValue?: Subscription | null,
	entity: NonNullable<Subscription['entity']>,
	entityId: number,
	type?: 'button' | 'dropdown',
}>(), {
	modelValue: null,
	type: 'button',
})

const emit = defineEmits<{
	'toggle': [subscribed: boolean]
}>()

const {t} = useI18n({useScope: 'global'})

const isInherited = computed(() => props.modelValue !== null &&
	(props.modelValue.entity !== props.entity || props.modelValue.entity_id !== props.entityId))

const tooltipText = computed(() => {
	if (isInherited.value) {
		return props.entity === 'task'
			? t('task.subscription.subscribedTaskThroughProject')
			: t('task.subscription.subscribedProjectThroughParentProject')
	}

	if (props.entity === 'task') {
		return props.modelValue !== null
			? t('task.subscription.subscribedTask')
			: t('task.subscription.notSubscribedTask')
	}

	return props.modelValue !== null
		? t('task.subscription.subscribedProject')
		: t('task.subscription.notSubscribedProject')
})

const buttonText = computed(() => props.modelValue ? t('task.subscription.unsubscribe') : t('task.subscription.subscribe'))
const iconName = computed<IconProp>(() => props.modelValue ? ['far', 'bell-slash'] : 'bell')
</script>
