<template>
	<dropdown>
		<template v-if="isSavedFilter">
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.edit`, params: { listId: list.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.delete`, params: { listId: list.id } }"
				icon="trash-alt"
			>
				{{ $t('misc.delete') }}
			</dropdown-item>
		</template>
		<template v-else-if="list.isArchived">
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.archive`, params: { listId: list.id } }"
				icon="archive"
			>
				{{ $t('menu.unarchive') }}
			</dropdown-item>
		</template>
		<template v-else>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.edit`, params: { listId: list.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.background`, params: { listId: list.id } }"
				v-if="backgroundsEnabled"
				icon="image"
			>
				{{ $t('menu.setBackground') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.share`, params: { listId: list.id } }"
				icon="share-alt"
			>
				{{ $t('menu.share') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.duplicate`, params: { listId: list.id } }"
				icon="paste"
			>
				{{ $t('menu.duplicate') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.archive`, params: { listId: list.id } }"
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
				:to="{ name: `${listRoutePrefix}.settings.delete`, params: { listId: list.id } }"
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
import Dropdown from '@/components/misc/dropdown'
import DropdownItem from '@/components/misc/dropdown-item'
import TaskSubscription from '@/components/misc/subscription'

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
			return this.$store.state.config.enabledBackgroundProviders.length > 0
		},
		listRoutePrefix() {
			let name = 'list'

			if (this.$route.name !== null && this.$route.name.startsWith('list.')) {
				name = this.$route.name
			}

			if (this.isSavedFilter) {
				name = name.replace('list.', 'filter.')
			}

			return name
		},
		isSavedFilter() {
			return getSavedFilterIdFromListId(this.list.id) > 0
		},
	},
}
</script>
