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
					:class="{'is-active': viewName === 'list'}"
					:to="{ name: 'list.list',   params: { listId } }">
					{{ $t('list.list.title') }}
				</router-link>
				<router-link
					v-shortcut="'g g'"
					:title="$t('keyboardShortcuts.list.switchToGanttView')"
					:class="{'is-active': viewName === 'gantt'}"
					:to="{ name: 'list.gantt',  params: { listId } }">
					{{ $t('list.gantt.title') }}
				</router-link>
				<router-link
					v-shortcut="'g t'"
					:title="$t('keyboardShortcuts.list.switchToTableView')"
					:class="{'is-active': viewName === 'table'}"
					:to="{ name: 'list.table',  params: { listId } }">
					{{ $t('list.table.title') }}
				</router-link>
				<router-link
					v-shortcut="'g k'"
					:title="$t('keyboardShortcuts.list.switchToKanbanView')"
					:class="{'is-active': viewName === 'kanban'}"
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

		<slot v-if="loadedListId"/>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'
import {useRoute} from 'vue-router'

import Message from '@/components/misc/message.vue'

import ListModel from '@/models/list'
import ListService from '@/services/list'

import {BACKGROUND, CURRENT_LIST} from '@/store/mutation-types'

import {getListTitle} from '@/helpers/getListTitle'
import {saveListToHistory} from '@/modules/listHistory'
import {useTitle} from '@/composables/useTitle'
import {useStore} from 'vuex'

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
const store = useStore()

const listService = ref(new ListService())
const loadedListId = ref(0)

const currentList = computed(() => {
	return typeof store.state.currentList === 'undefined' ? {
		id: 0,
		title: '',
		isArchived: false,
		maxRight: null,
	} : store.state.currentList
})

// watchEffect would be called every time the prop would get a value assigned, even if that value was the same as before.
// This resulted in loading and setting the list multiple times, even when navigating away from it.
// This caused wired bugs where the list background would be set on the home page but only right after setting a new 
// list background and then navigating to home. It also highlighted the list in the menu and didn't allow changing any
// of it, most likely due to the rights not being properly populated.
watch(
	() => props.listId,
	(listId, prevListId) => {
		loadList(listId)
	},
	{
		immediate: true,
	}
)

// call the method again if the listId changes
watchEffect(() => loadList(props.listId))

useTitle(() => currentList.value.id ? getListTitle(currentList.value) : '')

async function loadList(listIdToLoad: number) {
	const listData = {id: listIdToLoad}
	saveListToHistory(listData)

	// This invalidates the loaded list at the kanban board which lets it reload its content when
	// switched to it. This ensures updates done to tasks in the gantt or list views are consistently
	// shown in all views while preventing reloads when closing a task popup.
	// We don't do this for the table view because that does not change tasks.
	// FIXME: remove this
	if (
		props.viewName === 'list.list' ||
		props.viewName === 'list.gantt'
	) {
		store.commit('kanban/setListId', 0)
	}

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
		return
	}

	console.debug(`Loading list, props.viewName = ${props.viewName}, $route.params =`, route.params, `, loadedListId = ${loadedListId.value}, currentList = `, currentList.value)

	// Put set the current list to the one we're about to load so that the title is already shown at the top
	loadedListId.value = 0
	const listFromStore = store.getters['lists/getListById'](listData.id)
	if (listFromStore !== null) {
		store.commit(BACKGROUND, null)
		store.commit(CURRENT_LIST, listFromStore)
	}

	// We create an extra list object instead of creating it in list.value because that would trigger a ui update which would result in bad ux.
	const list = new ListModel(listData)
	try {
		const loadedList = await listService.value.get(list)
		await store.dispatch(CURRENT_LIST, loadedList)
	} finally {
		loadedListId.value = props.listId
	}
}
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