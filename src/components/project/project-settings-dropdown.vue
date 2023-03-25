<template>
	<dropdown>
		<template #trigger="triggerProps">
			<slot name="trigger" v-bind="triggerProps">
				<BaseButton class="dropdown-trigger" @click="triggerProps.toggleOpen">
					<icon icon="ellipsis-h" class="icon"/>
				</BaseButton>
			</slot>
		</template>

		<template v-if="isSavedFilter(project)">
			<dropdown-item
				:to="{ name: 'filter.settings.edit', params: { projectId: project.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'filter.settings.delete', params: { projectId: project.id } }"
				icon="trash-alt"
			>
				{{ $t('misc.delete') }}
			</dropdown-item>
		</template>

		<template v-else-if="project.isArchived">
			<dropdown-item
				:to="{ name: 'project.settings.archive', params: { projectId: project.id } }"
				icon="archive"
			>
				{{ $t('menu.unarchive') }}
			</dropdown-item>
		</template>
		<template v-else>
			<dropdown-item
				:to="{ name: 'project.settings.edit', params: { projectId: project.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</dropdown-item>
			<dropdown-item
				v-if="backgroundsEnabled"
				:to="{ name: 'project.settings.background', params: { projectId: project.id } }"
				icon="image"
			>
				{{ $t('menu.setBackground') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'project.settings.share', params: { projectId: project.id } }"
				icon="share-alt"
			>
				{{ $t('menu.share') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'project.settings.duplicate', params: { projectId: project.id } }"
				icon="paste"
			>
				{{ $t('menu.duplicate') }}
			</dropdown-item>
			<dropdown-item
				:to="{ name: 'project.settings.archive', params: { projectId: project.id } }"
				icon="archive"
			>
				{{ $t('menu.archive') }}
			</dropdown-item>
			<Subscription
				class="has-no-shadow"
				:is-button="false"
				entity="project"
				:entity-id="project.id"
				:model-value="project.subscription"
				@update:model-value="setSubscriptionInStore"
				type="dropdown"
			/>
			<dropdown-item
				:to="{ name: 'project.settings.delete', params: { projectId: project.id } }"
				icon="trash-alt"
				class="has-text-danger"
			>
				{{ $t('menu.delete') }}
			</dropdown-item>
		</template>
	</dropdown>
</template>

<script setup lang="ts">
import {ref, computed, watchEffect, type PropType} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Dropdown from '@/components/misc/dropdown.vue'
import DropdownItem from '@/components/misc/dropdown-item.vue'
import Subscription from '@/components/misc/subscription.vue'
import type {IProject} from '@/modelTypes/IProject'
import type {ISubscription} from '@/modelTypes/ISubscription'

import {isSavedFilter} from '@/services/savedFilter'
import {useConfigStore} from '@/stores/config'
import {useProjectStore} from '@/stores/projects'

const props = defineProps({
	project: {
		type: Object as PropType<IProject>,
		required: true,
	},
})

const projectStore = useProjectStore()
const subscription = ref<ISubscription | null>(null)
watchEffect(() => {
	subscription.value = props.project.subscription ?? null
})

const configStore = useConfigStore()
const backgroundsEnabled = computed(() => configStore.enabledBackgroundProviders?.length > 0)

function setSubscriptionInStore(sub: ISubscription) {
	subscription.value = sub
	const updatedProject = {
		...props.project,
		subscription: sub,
	}
	projectStore.setProject(updatedProject)
}
</script>