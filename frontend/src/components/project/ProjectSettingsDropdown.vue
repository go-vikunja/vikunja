<template>
	<Dropdown>
		<template #trigger="triggerProps">
			<slot
				name="trigger"
				v-bind="triggerProps"
			>
				<BaseButton
					class="dropdown-trigger"
					:aria-expanded="triggerProps.open"
					@click="triggerProps.toggleOpen"
				>
					<span class="is-sr-only">{{ $t('project.openSettingsMenu') }}</span>
					<Icon
						icon="ellipsis-h"
						class="icon"
					/>
				</BaseButton>
			</slot>
		</template>

		<template v-if="isSavedFilterProject(project)">
			<DropdownItem
				:to="{ name: 'filter.settings.edit', params: { projectId: project.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</DropdownItem>
			<DropdownItem
				:to="{ name: 'project.settings.views', params: { projectId: project.id } }"
				icon="eye"
			>
				{{ $t('menu.views') }}
			</DropdownItem>
			<slot name="before-delete" />
			<DropdownItem
				:to="{ name: 'filter.settings.delete', params: { projectId: project.id } }"
				icon="trash-alt"
				class="has-text-danger"
			>
				{{ $t('misc.delete') }}
			</DropdownItem>
		</template>

		<template v-else-if="project.is_archived">
			<DropdownItem
				:to="{ name: 'project.settings.archive', params: { projectId: project.id } }"
				icon="archive"
			>
				{{ $t('menu.unarchive') }}
			</DropdownItem>
		</template>
		<template v-else>
			<DropdownItem
				:to="{ name: 'project.settings.edit', params: { projectId: project.id } }"
				icon="pen"
			>
				{{ $t('menu.edit') }}
			</DropdownItem>
			<DropdownItem
				:to="{ name: 'project.settings.views', params: { projectId: project.id } }"
				icon="eye"
			>
				{{ $t('menu.views') }}
			</DropdownItem>
			<DropdownItem
				v-if="backgroundsEnabled"
				:to="{ name: 'project.settings.background', params: { projectId: project.id } }"
				icon="image"
			>
				{{ $t('menu.setBackground') }}
			</DropdownItem>
			<DropdownItem
				:to="{ name: 'project.settings.share', params: { projectId: project.id } }"
				icon="share-alt"
			>
				{{ $t('menu.share') }}
			</DropdownItem>
			<DropdownItem
				:to="{ name: 'project.settings.duplicate', params: { projectId: project.id } }"
				icon="paste"
			>
				{{ $t('menu.duplicate') }}
			</DropdownItem>
			<DropdownItem
				v-tooltip="isDefaultProject ? $t('menu.cantArchiveIsDefault') : ''"
				:to="{ name: 'project.settings.archive', params: { projectId: project.id } }"
				icon="archive"
				:disabled="isDefaultProject"
			>
				{{ $t('menu.archive') }}
			</DropdownItem>
			<Subscription
				class="has-no-shadow"
				entity="project"
				:entity-id="project.id"
				:model-value="subscription"
				type="dropdown"
				@update:modelValue="setSubscriptionInStore"
			/>
			<DropdownItem
				:to="{ name: 'project.settings.webhooks', params: { projectId: project.id } }"
				icon="bolt"
			>
				{{ $t('project.webhooks.title') }}
			</DropdownItem>
			<DropdownItem
				:to="{ name: 'project.createFromParent', params: { parentProjectId: project.id } }"
				icon="layer-group"
			>
				{{ $t('menu.createProject') }}
			</DropdownItem>
			<slot name="before-delete" />
			<DropdownItem
				v-if="forceAllActions || project.max_permission === PERMISSIONS.ADMIN"
				v-tooltip="isDefaultProject ? $t('menu.cantDeleteIsDefault') : ''"
				:to="{ name: 'project.settings.delete', params: { projectId: project.id } }"
				icon="trash-alt"
				class="has-text-danger"
				:disabled="isDefaultProject"
			>
				{{ $t('menu.delete') }}
			</DropdownItem>
		</template>
	</Dropdown>
</template>

<script setup lang="ts">
import {computed, ref, watchEffect} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import Subscription from '@/components/misc/Subscription.vue'
import type {Project, Subscription as ProjectSubscription} from '@/client/generated'
import type {ISubscription} from '@/modelTypes/ISubscription'
import SubscriptionModel from '@/models/subscription'

import {isSavedFilterProject} from '@/client/queries/projects'
import {useConfigStore} from '@/stores/config'
import {useProjectNavigation} from '@/composables/useProjectNavigation'
import {useAuthStore} from '@/stores/auth'
import {PERMISSIONS} from '@/constants/permissions'

const props = withDefaults(defineProps<{
	project: Project & Required<Pick<Project, 'id'>>
	forceAllActions?: boolean
}>(), {
	forceAllActions: false,
})

const projectStore = useProjectNavigation()
const subscription = ref<ISubscription | null>(null)
watchEffect(() => {
	const value = props.project.subscription
	subscription.value = value
		? new SubscriptionModel({
			id: value.id,
			entity: value.entity,
			entityId: value.entity_id,
			created: value.created ? new Date(value.created) : undefined,
		})
		: null
})

const configStore = useConfigStore()
const backgroundsEnabled = computed(() => configStore.enabledBackgroundProviders?.length > 0)

function toProjectSubscription(sub: ISubscription | null): ProjectSubscription | undefined {
	return sub
		? {
			id: sub.id,
			entity: sub.entity as ProjectSubscription['entity'],
			entity_id: sub.entityId,
			created: sub.created?.toISOString(),
		}
		: undefined
}

function setSubscriptionInStore(sub: ISubscription | null) {
	subscription.value = sub
	const updatedProject = {
		...props.project,
		subscription: toProjectSubscription(sub),
	}
	projectStore.setProject(updatedProject)
}

const authStore = useAuthStore()
const isDefaultProject = computed(() => props.project?.id === authStore.settings.defaultProjectId)
</script>
