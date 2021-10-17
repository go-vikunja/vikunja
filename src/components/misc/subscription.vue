<template>
	<x-button
		type="secondary"
		:icon="icon"
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:disabled="disabled || null"
		v-if="isButton"
	>
		{{ buttonText }}
	</x-button>
	<a
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:class="{'is-disabled': disabled}"
		v-else
	>
		<span class="icon">
			<icon :icon="icon"/>
		</span>
		{{ buttonText }}
	</a>
</template>

<script>
import SubscriptionService from '@/services/subscription'
import SubscriptionModel from '@/models/subscription'

export default {
	name: 'task-subscription',
	data() {
		return {
			subscriptionService: new SubscriptionService(),
		}
	},
	props: {
		entity: {
			required: true,
			type: String,
		},
		subscription: {
			required: true,
		},
		entityId: {
			required: true,
		},
		isButton: {
			type: Boolean,
			default: true,
		},
	},
	emits: ['change'],
	computed: {
		tooltipText() {
			if (this.disabled) {
				return this.$t('task.subscription.subscribedThroughParent', {
					entity: this.entity,
					parent: this.subscription.entity,
				})
			}

			return this.subscription !== null ?
				this.$t('task.subscription.subscribed', {entity: this.entity}) :
				this.$t('task.subscription.notSubscribed', {entity: this.entity})
		},
		buttonText() {
			return this.subscription !== null ? this.$t('task.subscription.unsubscribe') : this.$t('task.subscription.subscribe')
		},
		icon() {
			return this.subscription !== null ? ['far', 'bell-slash'] : 'bell'
		},
		disabled() {
			if (this.subscription === null) {
				return false
			}

			return this.subscription.entity !== this.entity
		},
	},
	methods: {
		changeSubscription() {
			if (this.disabled) {
				return
			}

			if (this.subscription === null) {
				this.subscribe()
			} else {
				this.unsubscribe()
			}
		},
		async subscribe() {
			const subscription = new SubscriptionModel({
				entity: this.entity,
				entityId: this.entityId,
			})
			await this.subscriptionService.create(subscription)
			this.$emit('change', subscription)
			this.$message.success({message: this.$t('task.subscription.subscribeSuccess', {entity: this.entity})})
		},
		async unsubscribe() {
			const subscription = new SubscriptionModel({
				entity: this.entity,
				entityId: this.entityId,
			})
			await this.subscriptionService.delete(subscription)
			this.$emit('change', null)
			this.$message.success({message: this.$t('task.subscription.unsubscribeSuccess', {entity: this.entity})})
		},
	},
}
</script>
