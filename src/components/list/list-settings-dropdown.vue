<template>
	<dropdown>
		<template v-if="isSavedFilter">
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.edit`, params: { listId: list.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.delete`, params: { listId: list.id } }"
				icon="trash-alt"
			>
				{{ $t('misc.delete') }}
			</dropdown-item>
		</template>
		<template v-else-if="list.isArchived">
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.archive`, params: { listId: list.id } }"
				icon="archive"
			>
				{{ $t('menu.unarchive') }}
			</dropdown-item>
		</template>
		<template v-else>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.edit`, params: { listId: list.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.background`, params: { listId: list.id } }"
				v-if="backgroundsEnabled"
				icon="image"
			>
				{{ $t('menu.setBackground') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.share`, params: { listId: list.id } }"
				icon="share-alt"
			>
				{{ $t('menu.share') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.duplicate`, params: { listId: list.id } }"
				icon="paste"
			>
				{{ $t('menu.duplicate') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: `${listRoutePrefix}.archive`, params: { listId: list.id } }"
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
				:to="{ name: `${listRoutePrefix}.delete`, params: { listId: list.id } }"
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
			return this.$store.state.config.enabledBackgroundProviders !== null && this.$store.state.config.enabledBackgroundProviders.length > 0
		},
		listRoutePrefix() {
			let name = 'list'


			if (this.$route.name !== null && this.$route.name.startsWith('list.')) {
				// HACK: we should implement a better routing for the modals
				const settingsRoutes = ['edit', 'delete', 'archive', 'background', 'share', 'duplicate']
				const suffix = settingsRoutes.find((route) => this.$route.name.endsWith(`.settings.${route}`))
				name = this.$route.name.replace(`.settings.${suffix}`,'')
			}

			if (this.isSavedFilter) {
				name = name.replace('list.', 'filter.')
			}

			return `${name}.settings`
		},
		isSavedFilter() {
			return getSavedFilterIdFromListId(this.list.id) > 0
		},
	},
}
</script>
