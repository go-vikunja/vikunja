<template>
	<x-button
		v-if="isButton"
		variant="secondary"
		:icon="iconName"
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:disabled="disabled || null"
	>
		{{ buttonText }}
	</x-button>
	<BaseButton
		v-else
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:class="{'is-disabled': disabled}"
	>
		<span class="icon">
			<icon :icon="iconName"/>
		</span>
		{{ buttonText }}
	</BaseButton>
</template>

<script lang="ts" setup>
import {computed, shallowRef} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'

import SubscriptionService from '@/services/subscription'
import SubscriptionModel from '@/models/subscription'

import {success} from '@/message'

interface Props {
  entity: string
  entityId: number
  subscription: SubscriptionModel | null
  isButton?: boolean
}

const props = withDefaults(defineProps<Props>(), {
	isButton: true,
	subscription: null,
})

const subscriptionEntity = computed<string | null>(() => props.subscription?.entity ?? null)

const emit = defineEmits(['change'])

const subscriptionService = shallowRef(new SubscriptionService())

const {t} = useI18n()
const tooltipText = computed(() => {
	if (disabled.value) {
		return t('task.subscription.subscribedThroughParent', {
			entity: props.entity,
			parent: subscriptionEntity.value,
		})
	}

	return props.subscription !== null ?
		t('task.subscription.subscribed', {entity: props.entity}) :
		t('task.subscription.notSubscribed', {entity: props.entity})
})

const buttonText = computed(() => props.subscription !== null ? t('task.subscription.unsubscribe') : t('task.subscription.subscribe'))
const iconName = computed(() => props.subscription !== null ? ['far', 'bell-slash'] : 'bell')
const disabled = computed(() => {
	if (props.subscription === null) {
		return false
	}

	return subscriptionEntity.value !== props.entity
})

function changeSubscription() {
	if (disabled.value) {
		return
	}

	if (props.subscription === null) {
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
	await subscriptionService.value.create(subscription)
	emit('change', subscription)
	success({message: t('task.subscription.subscribeSuccess', {entity: props.entity})})
}

async function unsubscribe() {
	const subscription = new SubscriptionModel({
		entity: props.entity,
		entityId: props.entityId,
	})
	await subscriptionService.value.delete(subscription)
	emit('change', null)
	success({message: t('task.subscription.unsubscribeSuccess', {entity: props.entity})})
}
</script>
