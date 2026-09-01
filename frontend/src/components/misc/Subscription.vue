<template>
	<XButton
		v-if="type === 'button'"
		v-tooltip="tooltipText"
		variant="secondary"
		:icon="iconName"
		@click="changeSubscription"
	>
		{{ buttonText }}
	</XButton>
	<DropdownItem
		v-else-if="type === 'dropdown'"
		v-tooltip="tooltipText"
		:icon="iconName"
		@click="changeSubscription"
	>
		{{ buttonText }}
	</DropdownItem>
</template>

<script lang="ts" setup>
import {computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import DropdownItem from '@/components/misc/DropdownItem.vue'

import SubscriptionService from '@/services/subscription'
import SubscriptionModel from '@/models/subscription'
import type {ISubscription} from '@/modelTypes/ISubscription'

import {success} from '@/message'
import type { IconProp } from '@fortawesome/fontawesome-svg-core'

const props = withDefaults(defineProps<{
	modelValue: ISubscription | null,
	entity: ISubscription['entity'],
	entityId: number,
	type?: 'button' | 'dropdown',
}>(), {
	type: 'button',
})

const emit = defineEmits<{
	'update:modelValue': [subscription: ISubscription | null]
}>()

const subscriptionService = shallowReactive(new SubscriptionService())

const {t} = useI18n({useScope: 'global'})

const isInherited = computed(() => props.modelValue !== null &&
	(props.modelValue.entity !== props.entity || props.modelValue.entityId !== props.entityId))

const tooltipText = computed(() => {
	if (isInherited.value) {
		return props.entity === 'task'
			? t('task.subscription.subscribedTaskThroughProject')
			: t('task.subscription.subscribedProjectThroughParentProject')
	}

	switch (props.entity) {
		case 'project':
			return props.modelValue !== null ?
				t('task.subscription.subscribedProject') :
				t('task.subscription.notSubscribedProject')
		case 'task':
			return props.modelValue !== null ?
				t('task.subscription.subscribedTask') :
				t('task.subscription.notSubscribedTask')
	}

	return ''
})

const buttonText = computed(() => props.modelValue ? t('task.subscription.unsubscribe') : t('task.subscription.subscribe'))
const iconName = computed<IconProp>(() => props.modelValue ? ['far', 'bell-slash'] : 'bell')

function changeSubscription() {
	return props.modelValue === null
		? subscribe()
		: unsubscribe()
}

async function subscribe() {
	const subscription = new SubscriptionModel({
		entity: props.entity,
		entityId: props.entityId,
	})
	await subscriptionService.create(subscription)
	emit('update:modelValue', subscription)

	let message = ''
	switch (props.entity) {
		case 'project':
			message = t('task.subscription.subscribeSuccessProject')
			break
		case 'task':
			message = t('task.subscription.subscribeSuccessTask')
			break
	}
	success({message})
}

async function unsubscribe() {
	const subscription = new SubscriptionModel({
		entity: props.entity,
		entityId: props.entityId,
	})
	await subscriptionService.delete(subscription)
	emit('update:modelValue', null)

	let message = ''
	switch (props.entity) {
		case 'project':
			message = t('task.subscription.unsubscribeSuccessProject')
			break
		case 'task':
			message = t('task.subscription.unsubscribeSuccessTask')
			break
	}
	success({message})
}
</script>
