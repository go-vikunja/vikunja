<template>
	<dropdown>
		<template v-if="isSavedFilter">
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.edit`, params: { listId: list.id } }"
				icon="pen"
			>
				Edit
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.delete`, params: { listId: list.id } }"
				icon="trash-alt"
			>
				Delete
			</dropdown-item>
		</template>
		<template v-else-if="list.isArchived">
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.archive`, params: { listId: list.id } }"
				icon="archive"
			>
				Un-Archive
			</dropdown-item>
		</template>
		<template v-else>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.edit`, params: { listId: list.id } }"
				icon="pen"
			>
				Edit
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.background`, params: { listId: list.id } }"
				v-if="backgroundsEnabled"
				icon="image"
			>
				Set background
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.share`, params: { listId: list.id } }"
				icon="share-alt"
			>
				Share
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.duplicate`, params: { listId: list.id } }"
				icon="paste"
			>
				Duplicate
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.settings.archive`, params: { listId: list.id } }"
				icon="archive"
			>
				Archive
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
				Delete
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

			if (this.$route.name.startsWith('list.')) {
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
