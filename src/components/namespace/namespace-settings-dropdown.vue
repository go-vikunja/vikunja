<template>
	<dropdown>
		<template v-if="namespace.isArchived">
			<dropdown-item
				:to="{ name: 'namespace.settings.archive', params: { id: namespace.id } }"
				icon="archive"
			>
				Un-Archive
			</dropdown-item>
		</template>
		<template v-else>
			<dropdown-item
				:to="{ name: 'namespace.settings.edit', params: { id: namespace.id } }"
				icon="pen"
			>
				Edit
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'namespace.settings.share', params: { id: namespace.id } }"
				icon="share-alt"
			>
				Share
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'list.create', params: { id: namespace.id } }"
				icon="plus"
			>
				New list
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'namespace.settings.archive', params: { id: namespace.id } }"
				icon="archive"
			>
				Archive
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'namespace.settings.delete', params: { id: namespace.id } }"
				icon="trash-alt"
				class="has-text-danger"
			>
				Delete
			</dropdown-item>
		</template>
	</dropdown>
</template>

<script>
import Dropdown from '@/components/misc/dropdown'
import DropdownItem from '@/components/misc/dropdown-item'

export default {
	name: 'namespace-settings-dropdown',
	components: {
		DropdownItem,
		Dropdown,
	},
	props: {
		namespace: {
			required: true,
		},
	},
}
</script>
