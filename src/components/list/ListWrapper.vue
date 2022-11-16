<template>
	<div
		:class="{ 'is-loading': listService.loading, 'is-archived': currentList.isArchived}"
		class="loader-container"
	>
		<div class="switch-view-container">
			<div class="switch-view">
				<BaseButton
					v-shortcut="'g l'"
					:title="$t('keyboardShortcuts.list.switchToListView')"
					class="switch-view-button"
					:class="{'is-active': viewName === 'list'}"
					:to="{ name: 'list.list',   params: { listId } }"
				>
					{{ $t('list.list.title') }}
				</BaseButton>
				<BaseButton
					v-shortcut="'g g'"
					:title="$t('keyboardShortcuts.list.switchToGanttView')"
					class="switch-view-button"
					:class="{'is-active': viewName === 'gantt'}"
					:to="{ name: 'list.gantt',  params: { listId } }"
				>
					{{ $t('list.gantt.title') }}
				</BaseButton>
				<BaseButton
					v-shortcut="'g t'"
					:title="$t('keyboardShortcuts.list.switchToTableView')"
					class="switch-view-button"
					:class="{'is-active': viewName === 'table'}"
					:to="{ name: 'list.table',  params: { listId } }"
				>
					{{ $t('list.table.title') }}
				</BaseButton>
				<BaseButton
					v-shortcut="'g k'"
					:title="$t('keyboardShortcuts.list.switchToKanbanView')"
					class="switch-view-button"
					:class="{'is-active': viewName === 'kanban'}"
					:to="{ name: 'list.kanban', params: { listId } }"
				>
					{{ $t('list.kanban.title') }}
				</BaseButton>
			</div>
			<slot name="header" />
		</div>
		<CustomTransition name="fade">
			<Message variant="warning" v-if="currentList.isArchived" class="mb-4">
				{{ $t('list.archived') }}
			</Message>
		</CustomTransition>

		<slot v-if="loadedListId"/>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'
import {useRoute} from 'vue-router'

import BaseButton from '@/components/base/BaseButton.vue'
import Message from '@/components/misc/message.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import ListModel from '@/models/list'
import ListService from '@/services/list'

import {getListTitle} from '@/helpers/getListTitle'
import {saveListToHistory} from '@/modules/listHistory'
import {useTitle} from '@/composables/useTitle'

import {useBaseStore} from '@/stores/base'
import {useListStore} from '@/stores/lists'

const props = defineProps({
	listId: {
		type: Number,
		required: true,
	},
	viewName: {
		type: String,
		required: true,
	},
})

const route = useRoute()

const baseStore = useBaseStore()
const listStore = useListStore()
const listService = ref(new ListService())
const loadedListId = ref(0)

const currentList = computed(() => {
	return typeof baseStore.currentList === 'undefined' ? {
		id: 0,
		title: '',
		isArchived: false,
		maxRight: null,
	} : baseStore.currentList
})
useTitle(() => currentList.value.id ? getListTitle(currentList.value) : '')

// watchEffect would be called every time the prop would get a value assigned, even if that value was the same as before.
// This resulted in loading and setting the list multiple times, even when navigating away from it.
// This caused wired bugs where the list background would be set on the home page but only right after setting a new 
// list background and then navigating to home. It also highlighted the list in the menu and didn't allow changing any
// of it, most likely due to the rights not being properly populated.
watch(
	() => props.listId,
	// loadList
	async (listIdToLoad: number) => {
		const listData = {id: listIdToLoad}
		saveListToHistory(listData)

		// Don't load the list if we either already loaded it or aren't dealing with a list at all currently and
		// the currently loaded list has the right set.
		if (
			(
				listIdToLoad === loadedListId.value ||
				typeof listIdToLoad === 'undefined' ||
				listIdToLoad === currentList.value.id
			)
			&& typeof currentList.value !== 'undefined' && currentList.value.maxRight !== null
		) {
			loadedListId.value = props.listId
			return
		}

		console.debug(`Loading list, props.viewName = ${props.viewName}, $route.params =`, route.params, `, loadedListId = ${loadedListId.value}, currentList = `, currentList.value)

		// Set the current list to the one we're about to load so that the title is already shown at the top
		loadedListId.value = 0
		const listFromStore = listStore.getListById(listData.id)
		if (listFromStore !== null) {
			baseStore.setBackground(null)
			baseStore.setBlurHash(null)
			baseStore.handleSetCurrentList({list: listFromStore})
		}

		// We create an extra list object instead of creating it in list.value because that would trigger a ui update which would result in bad ux.
		const list = new ListModel(listData)
		try {
			const loadedList = await listService.value.get(list)
			baseStore.handleSetCurrentList({list: loadedList})
		} finally {
			loadedListId.value = props.listId
		}
	},
	{immediate: true},
)
</script>

<style lang="scss" scoped>
.switch-view-container {
  @media screen and (max-width: $tablet) {
    display: flex;
    justify-content: center;
    flex-direction: column;
  }
}

.switch-view {
  background: var(--white);
  display: inline-flex;
  border-radius: $radius;
  font-size: .75rem;
  box-shadow: var(--shadow-sm);
  height: $switch-view-height;
  margin: 0 auto 1rem;
  padding: .5rem;
}

.switch-view-button {
	padding: .25rem .5rem;
	display: block;
	border-radius: $radius;
	transition: all 100ms;

	&:not(:last-child) {
		margin-right: .5rem;
	}

	&:hover {
		color: var(--switch-view-color);
		background: var(--primary);
	}

	&.is-active {
		color: var(--switch-view-color);
		background: var(--primary);
		font-weight: bold;
		box-shadow: var(--shadow-xs);
	}
}

// FIXME: this should be in notification and set via a prop
.is-archived .notification.is-warning {
  margin-bottom: 1rem;
}
</style>