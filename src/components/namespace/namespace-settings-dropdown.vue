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
			<task-subscription
				class="dropdown-item has-no-shadow"
				:is-button="false"
				entity="namespace"
				:entity-id="namespace.id"
				:subscription="subscription"
				@change="sub => subscription = sub"
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
import {ref, onMounted} from 'vue'

import Dropdown from '@/components/misc/dropdown.vue'
import DropdownItem from '@/components/misc/dropdown-item.vue'
import TaskSubscription from '@/components/misc/subscription.vue'

const props = defineProps({
	namespace: {
		type: Object, // NamespaceModel
		required: true,
	},
})

const subscription = ref(null)
onMounted(() => {
	subscription.value = props.namespace.subscription
})
</script>
