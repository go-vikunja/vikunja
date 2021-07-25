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
				:to="{ name: 'namespace.settings.share', params: { id: namespace.id } }"
				icon="share-alt"
			>
				{{ $t('menu.share') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'list.create', params: { id: namespace.id } }"
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

<script>
import Dropdown from '@/components/misc/dropdown.vue'
import DropdownItem from '@/components/misc/dropdown-item.vue'
import TaskSubscription from '@/components/misc/subscription.vue'

export default {
	name: 'namespace-settings-dropdown',
	data() {
		return {
			subscription: null,
		}
	},
	components: {
		DropdownItem,
		Dropdown,
		TaskSubscription,
	},
	props: {
		namespace: {
			required: true,
		},
	},
	mounted() {
		this.subscription = this.namespace.subscription
	},
}
</script>
