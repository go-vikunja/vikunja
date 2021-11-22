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
					:class="{'is-active': $route.name.includes('list.list')}"
					:to="{ name: 'list.list',   params: { listId: listId } }">
					{{ $t('list.list.title') }}
				</router-link>
				<router-link
					v-shortcut="'g g'"
					:title="$t('keyboardShortcuts.list.switchToGanttView')"
					:class="{'is-active': $route.name.includes('list.gantt')}"
					:to="{ name: 'list.gantt',  params: { listId: listId } }">
					{{ $t('list.gantt.title') }}
				</router-link>
				<router-link
					v-shortcut="'g t'"
					:title="$t('keyboardShortcuts.list.switchToTableView')"
					:class="{'is-active': $route.name.includes('list.table')}"
					:to="{ name: 'list.table',  params: { listId: listId } }">
					{{ $t('list.table.title') }}
				</router-link>
				<router-link
					v-shortcut="'g k'"
					:title="$t('keyboardShortcuts.list.switchToKanbanView')"
					:class="{'is-active': $route.name.includes('list.kanban')}"
					:to="{ name: 'list.kanban', params: { listId: listId } }">
					{{ $t('list.kanban.title') }}
				</router-link>
			</div>
		</div>
		<transition name="fade">
			<div class="notification is-warning" v-if="currentList.isArchived">
				{{ $t('list.archived') }}
			</div>
		</transition>

		<router-view/>
	</div>
</template>

<script>
import ListModel from '../../models/list'
import ListService from '../../services/list'
import {CURRENT_LIST} from '../../store/mutation-types'
import {getListView} from '../../helpers/saveListView'
import {saveListToHistory} from '../../modules/listHistory'

export default {
	data() {
		return {
			listService: new ListService(),
			listLoaded: 0,
		}
	},
	watch: {
		// call again the method if the route changes
		'$route.path': {
			handler: 'loadList',
			immediate: true,
		},
	},
	computed: {
		// Computed property to let "listId" always have a value
		listId() {
			return typeof this.$route.params.listId === 'undefined' ? 0 : this.$route.params.listId
		},
		background() {
			return this.$store.state.background
		},
		currentList() {
			return typeof this.$store.state.currentList === 'undefined' ? {
				id: 0,
				title: '',
				isArchived: false,
			} : this.$store.state.currentList
		},
	},
	methods: {
		replaceListView() {
			const savedListView = getListView(this.$route.params.listId)
			this.$router.replace({name: savedListView, params: {id: this.$route.params.listId}})
			console.debug('Replaced list view with', savedListView)
		},

		async loadList() {
			if (this.$route.name.includes('.settings.')) {
				return
			}

			const listData = {id: parseInt(this.$route.params.listId)}

			saveListToHistory(listData)

			this.setTitle(this.currentList.id ? this.getListTitle(this.currentList) : '')

			// This invalidates the loaded list at the kanban board which lets it reload its content when
			// switched to it. This ensures updates done to tasks in the gantt or list views are consistently
			// shown in all views while preventing reloads when closing a task popup.
			// We don't do this for the table view because that does not change tasks.
			if (
				this.$route.name === 'list.list' ||
				this.$route.name === 'list.gantt'
			) {
				this.$store.commit('kanban/setListId', 0)
			}

			// When clicking again on a list in the menu, there would be no list view selected which means no list
			// at all. Users will then have to click on the list view menu again which is quite confusing.
			if (this.$route.name === 'list.index') {
				return this.replaceListView()
			}

			// Don't load the list if we either already loaded it or aren't dealing with a list at all currently and
			// the currently loaded list has the right set.
			if (
				(
					this.$route.params.listId === this.listLoaded ||
					typeof this.$route.params.listId === 'undefined' ||
					this.$route.params.listId === this.currentList.id ||
					parseInt(this.$route.params.listId) === this.currentList.id
				)
				&& typeof this.currentList !== 'undefined' && this.currentList.maxRight !== null
			) {
				return
			}

			// Redirect the user to list view by default
			if (
				this.$route.name !== 'list.list' &&
				this.$route.name !== 'list.gantt' &&
				this.$route.name !== 'list.table' &&
				this.$route.name !== 'list.kanban'
			) {
				return this.replaceListView()
			}

			console.debug(`Loading list, $route.name = ${this.$route.name}, $route.params =`, this.$route.params, `, listLoaded = ${this.listLoaded}, currentList = `, this.currentList)

			// We create an extra list object instead of creating it in this.list because that would trigger a ui update which would result in bad ux.
			const list = new ListModel(listData)
			try {
				const loadedList = await this.listService.get(list)
				await this.$store.dispatch(CURRENT_LIST, loadedList)
				this.setTitle(this.getListTitle(loadedList))
			} finally {
				this.listLoaded = this.$route.params.listId
			}
		},
	},
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