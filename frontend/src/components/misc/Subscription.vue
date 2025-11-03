<template>
	<XButton
		v-if="type === 'button'"
		v-tooltip="tooltipText"
		variant="secondary"
		:icon="iconName"
		:disabled="disabled"
		@click="changeSubscription"
	>
		{{ buttonText }}
	</XButton>
	<DropdownItem
		v-else-if="type === 'dropdown'"
		v-tooltip="tooltipText"
		:disabled="disabled"
		:icon="iconName"
		@click="changeSubscription"
	>
		{{ buttonText }}
	</DropdownItem>
	<BaseButton
		v-else
		v-tooltip="tooltipText"
		:class="{'is-disabled': disabled}"
		:disabled="disabled"
		@click="changeSubscription"
	>
		<span class="icon">
			<Icon :icon="iconName" />
		</span>
		{{ buttonText }}
	</BaseButton>
</template>

<script lang="ts" setup>
import {computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
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
	isButton?: boolean,
	type?: 'button' | 'dropdown',
}>(), {
	isButton: true,
	type: 'button',
})

const emit = defineEmits<{
	'update:modelValue': [subscription: ISubscription | null]
}>()

const subscriptionEntity = computed<string | null>(() => props.modelValue?.entity ?? null)

const subscriptionService = shallowReactive(new SubscriptionService())

const {t} = useI18n({useScope: 'global'})

const tooltipText = computed(() => {
	if (disabled.value) {
		if (props.entity === 'task' && subscriptionEntity.value === 'project') {
			return t('task.subscription.subscribedTaskThroughParentProject')
		}

		return ''
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
const disabled = computed(() => props.modelValue && subscriptionEntity.value !== props.entity || false)

function changeSubscription() {
	if (disabled.value) {
		return
	}

	if (props.modelValue === null) {
		subscribe()
	} else {
		unsubscribe()
	}
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
