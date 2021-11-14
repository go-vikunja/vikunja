<template>
	<div
		:class="{ 'is-loading': listService.loading, 'is-archived': currentList.isArchived}"
		class="loader-container"
	>
		<div class="switch-view-container">
			<div class="switch-view">
				<router-link
					v-shortcut="'g l'"
					:title="$t('keyboardShortcuts.list.switchToListView')"
					:class="{'is-active': $route.name === 'list.list'}"
					:to="{ name: 'list.list',   params: { listId } }">
					{{ $t('list.list.title') }}
				</router-link>
				<router-link
					v-shortcut="'g g'"
					:title="$t('keyboardShortcuts.list.switchToGanttView')"
					:class="{'is-active': $route.name === 'list.gantt'}"
					:to="{ name: 'list.gantt',  params: { listId } }">
					{{ $t('list.gantt.title') }}
				</router-link>
				<router-link
					v-shortcut="'g t'"
					:title="$t('keyboardShortcuts.list.switchToTableView')"
					:class="{'is-active': $route.name === 'list.table'}"
					:to="{ name: 'list.table',  params: { listId } }">
					{{ $t('list.table.title') }}
				</router-link>
				<router-link
					v-shortcut="'g k'"
					:title="$t('keyboardShortcuts.list.switchToKanbanView')"
					:class="{'is-active': $route.name === 'list.kanban'}"
					:to="{ name: 'list.kanban', params: { listId } }">
					{{ $t('list.kanban.title') }}
				</router-link>
			</div>
			<slot name="header" />
		</div>
		<transition name="fade">
			<Message variant="warning" v-if="currentList.isArchived" class="mb-4">
				{{ $t('list.archived') }}
			</Message>
		</transition>

		<slot />
	</div>
</template>

<script setup>
import {ref, shallowRef, computed, watchEffect} from 'vue'
import {useRoute} from 'vue-router'

import Message from '@/components/misc/message'

import ListModel from '@/models/list'
import ListService from '@/services/list'

import {store} from '@/store'
import {CURRENT_LIST} from '@/store/mutation-types'

import {getListTitle} from '@/helpers/getListTitle'
import {saveListToHistory} from '@/modules/listHistory'
import { useTitle } from '@/composables/useTitle'

const route = useRoute()

const listService = shallowRef(new ListService())
const loadedListId = ref(0)

const currentList = computed(() => {
	return typeof store.state.currentList === 'undefined' ? {
		id: 0,
		title: '',
		isArchived: false,
	} : store.state.currentList
})

// Computed property to let "listId" always have a value
const listId = computed(() => typeof route.params.listId === 'undefined' ? 0 : parseInt(route.params.listId))
// call again the method if the listId changes
watchEffect(() => loadList(listId.value))

useTitle(() => currentList.value.id ? getListTitle(currentList.value) : '')

async function loadList(listId) {
	const listData = {id: listId}
	saveListToHistory(listData)

	// This invalidates the loaded list at the kanban board which lets it reload its content when
	// switched to it. This ensures updates done to tasks in the gantt or list views are consistently
	// shown in all views while preventing reloads when closing a task popup.
	// We don't do this for the table view because that does not change tasks.
	// FIXME: remove this
	if (
		route.name === 'list.list' ||
		route.name === 'list.gantt'
	) {
		store.commit('kanban/setListId', 0)
	}

	// Don't load the list if we either already loaded it or aren't dealing with a list at all currently and
	// the currently loaded list has the right set.
	if (
		(
			listId.value === loadedListId.value ||
			typeof listId.value === 'undefined' ||
			listId.value === currentList.value.id
		)
		&& typeof currentList.value !== 'undefined' && currentList.value.maxRight !== null
	) {
		return
	}

	console.debug(`Loading list, $route.name = ${route.name}, $route.params =`, route.params, `, loadedListId = ${loadedListId.value}, currentList = `, currentList.value)

	// We create an extra list object instead of creating it in list.value because that would trigger a ui update which would result in bad ux.
	const list = new ListModel(listData)
	try {
		const loadedList = await listService.value.get(list)
		await store.dispatch(CURRENT_LIST, loadedList)
	} finally {
		loadedListId.value = listId
	}
}
</script>

<style lang="scss" scoped>
.switch-view-container {
  @media screen and (max-width: $tablet) {
    display: flex;
    justify-content: center;
  }
}

.switch-view {
  background: var(--white);
  display: inline-flex;
  border-radius: $radius;
  font-size: .75rem;
  box-shadow: var(--shadow-sm);
  height: $switch-view-height;
  margin-bottom: 1rem;
  padding: .5rem;

  a {
    padding: .25rem .5rem;
    display: block;
    border-radius: $radius;

    transition: all 100ms;

    &:not(:last-child) {
      margin-right: .5rem;
    }

    &.is-active,
    &:hover {
      color: var(--switch-view-color);
    }

    &.is-active {
      background: var(--primary);
      font-weight: bold;
      box-shadow: var(--shadow-xs);
    }

    &:hover {
      background: var(--primary);
    }
  }
}

.is-archived .notification.is-warning {
  margin-bottom: 1rem;
}
</style>