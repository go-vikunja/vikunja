<template>
	<dropdown>
		<template v-if="isSavedFilter">
			<dropdown-item
				:to="{ name: 'filter.settings.edit', params: { listId: list.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'filter.settings.delete', params: { listId: list.id } }"
				icon="trash-alt"
			>
				{{ $t('misc.delete') }}
			</dropdown-item>
		</template>

		<template v-else-if="list.isArchived">
			<dropdown-item
				:to="{ name: 'list.settings.archive', params: { listId: list.id } }"
				icon="archive"
			>
				{{ $t('menu.unarchive') }}
			</dropdown-item>
		</template>
		<template v-else>
			<dropdown-item
				:to="{ name: 'list.settings.edit', params: { listId: list.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</dropdown-item>
			<dropdown-item
				v-if="backgroundsEnabled"
				:to="{ name: 'list.settings.background', params: { listId: list.id } }"
				icon="image"
			>
				{{ $t('menu.setBackground') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'list.settings.share', params: { listId: list.id } }"
				icon="share-alt"
			>
				{{ $t('menu.share') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'list.settings.duplicate', params: { listId: list.id } }"
				icon="paste"
			>
				{{ $t('menu.duplicate') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'list.settings.archive', params: { listId: list.id } }"
				icon="archive"
			>
				{{ $t('menu.archive') }}
			</dropdown-item>
			<task-subscription
				class="dropdown-item has-no-shadow"
				:is-button="false"
				entity="list"
				:entity-id="list.id"
				:subscription="subscription"
				@change="sub => subscription = sub"
			/>
			<dropdown-item
				:to="{ name: 'list.settings.delete', params: { listId: list.id } }"
				icon="trash-alt"
				class="has-text-danger"
			>
				{{ $t('menu.delete') }}
			</dropdown-item>
		</template>
	</dropdown>
</template>

<script>
import {getSavedFilterIdFromListId} from '@/helpers/savedFilter'
import Dropdown from '@/components/misc/dropdown.vue'
import DropdownItem from '@/components/misc/dropdown-item.vue'
import TaskSubscription from '@/components/misc/subscription.vue'

export default {
	name: 'list-settings-dropdown',
	data() {
		return {
			subscription: null,
		}
	},
	components: {
		TaskSubscription,
		DropdownItem,
		Dropdown,
	},
	props: {
		list: {
			required: true,
		},
	},
	mounted() {
		this.subscription = this.list.subscription
	},
	computed: {
		backgroundsEnabled() {
			return this.$store.state.config.enabledBackgroundProviders?.length > 0
		},
		isSavedFilter() {
			return getSavedFilterIdFromListId(this.list.id) > 0
		},
	},
}
</script>
