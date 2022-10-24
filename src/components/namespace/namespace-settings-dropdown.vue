<template>
	<dropdown>
		<template v-if="namespace.isArchived">
			<dropdown-item
				:to="{ name: 'namespace.settings.archive', params: { id: namespace.id } }"
				icon="archive"
			>
				{{ $t('menu.unarchive') }}
			</dropdown-item>
		</template>
		<template v-else>
			<dropdown-item
				:to="{ name: 'namespace.settings.edit', params: { id: namespace.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'namespace.settings.share', params: { namespaceId: namespace.id } }"
				icon="share-alt"
			>
				{{ $t('menu.share') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'list.create', params: { namespaceId: namespace.id } }"
				icon="plus"
			>
				{{ $t('menu.newList') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'namespace.settings.archive', params: { id: namespace.id } }"
				icon="archive"
			>
				{{ $t('menu.archive') }}
			</dropdown-item>
			<Subscription
				class="has-no-shadow"
				:is-button="false"
				entity="namespace"
				:entity-id="namespace.id"
				:model-value="subscription"
				@update:model-value="setSubscriptionInStore"
				type="dropdown"
			/>
			<dropdown-item
				:to="{ name: 'namespace.settings.delete', params: { id: namespace.id } }"
				icon="trash-alt"
				class="has-text-danger"
			>
				{{ $t('menu.delete') }}
			</dropdown-item>
		</template>
	</dropdown>
</template>

<script setup lang="ts">
import {ref, onMounted, type PropType} from 'vue'

import Dropdown from '@/components/misc/dropdown.vue'
import DropdownItem from '@/components/misc/dropdown-item.vue'
import Subscription from '@/components/misc/subscription.vue'
import type {INamespace} from '@/modelTypes/INamespace'
import type {ISubscription} from '@/modelTypes/ISubscription'
import {useNamespaceStore} from '@/stores/namespaces'

const props = defineProps({
	namespace: {
		type: Object as PropType<INamespace>,
		required: true,
	},
})

const namespaceStore = useNamespaceStore()

const subscription = ref<ISubscription | null>(null)
onMounted(() => {
	subscription.value = props.namespace.subscription
})

function setSubscriptionInStore(sub: ISubscription) {
	subscription.value = sub
	namespaceStore.setNamespaceById({
		...props.namespace,
		subscription: sub,
	})
}
</script>
