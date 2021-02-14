<template>
	<x-button
		type="secondary"
		:icon="icon"
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:disabled="disabled"
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
			subscriptionService: SubscriptionService,
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
	created() {
		this.subscriptionService = new SubscriptionService()
	},
	computed: {
		tooltipText() {
			if(this.disabled) {
				return `You can't unsubscribe here because you are subscribed to this ${this.entity} through its ${this.subscription.entity}.`
			}

			return this.subscription !== null ?
				`You are currently subscribed to this ${this.entity} and will receive notifications for changes.` :
				`You are not subscribed to this ${this.entity} and won't receive notifications for changes.`
		},
		buttonText() {
			return this.subscription !== null ? 'Unsubscribe' : 'Subscribe'
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
			if(this.disabled) {
				return
			}

			if (this.subscription === null) {
				this.subscribe()
			} else {
				this.unsubscribe()
			}
		},
		subscribe() {
			const subscription = new SubscriptionModel({
				entity: this.entity,
				entityId: this.entityId,
			})
			this.subscriptionService.create(subscription)
				.then(() => {
					this.$emit('change', subscription)
					this.success({message: `You are now subscribed to this ${this.entity}`}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		unsubscribe() {
			const subscription = new SubscriptionModel({
				entity: this.entity,
				entityId: this.entityId,
			})
			this.subscriptionService.delete(subscription)
				.then(() => {
					this.$emit('change', null)
					this.success({message: `You are now unsubscribed to this ${this.entity}`}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		}
	},
}
</script>
