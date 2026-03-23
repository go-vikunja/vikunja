<template>
	<Dropdown>
		<template #trigger="triggerProps">
			<slot
				name="trigger"
				v-bind="triggerProps"
			>
				<BaseButton
					class="dropdown-trigger"
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

		<template v-if="isSavedFilter(project)">
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
			<DropdownItem
				:to="{ name: 'filter.settings.delete', params: { projectId: project.id } }"
				icon="trash-alt"
				class="has-text-danger"
			>
				{{ $t('misc.delete') }}
			</DropdownItem>
		</template>

		<template v-else-if="project.isArchived">
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
				:is-button="false"
				entity="project"
				:entity-id="project.id"
				:model-value="project.subscription"
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
			<DropdownItem
				v-if="project.maxPermission === PERMISSIONS.ADMIN"
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
import type {IProject} from '@/modelTypes/IProject'
import type {ISubscription} from '@/modelTypes/ISubscription'

import {isSavedFilter} from '@/services/savedFilter'
import {useConfigStore} from '@/stores/config'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import {PERMISSIONS} from '@/constants/permissions'

const props = defineProps<{
	project: IProject
}>()

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

const authStore = useAuthStore()
const isDefaultProject = computed(() => props.project?.id === authStore.settings.defaultProjectId)
</script>
