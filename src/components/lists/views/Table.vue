<template>
	<div class="table-view loader-container" :class="{'is-loading': taskCollectionService.loading}">
		<div class="filter-container">
			<div class="items">
				<button class="button" @click="() => {showActiveColumnsFilter = !showActiveColumnsFilter; showTaskFilter = false}">
					<span class="icon is-small">
						<icon icon="th"/>
					</span>
					Columns
				</button>
				<button class="button" @click="() => {showTaskFilter = !showTaskFilter; showActiveColumnsFilter = false}">
					<span class="icon is-small">
						<icon icon="filter"/>
					</span>
					Filters
				</button>
			</div>
			<transition name="fade">
				<div class="card" v-if="showActiveColumnsFilter">
					<div class="card-content">
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.id">#</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.done">Done</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.title">Title</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.priority">Priority</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.labels">Labels</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.assignees">Assignees</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.dueDate">Due Date</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.startDate">Start Date</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.endDate">End Date</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.percentDone">% Done</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.created">Created</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.updated">Updated</fancycheckbox>
						<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.createdBy">Created By</fancycheckbox>
					</div>
				</div>
				<filters
						v-if="showTaskFilter"
						v-model="params"
						@change="loadTasks(1)"
				/>
			</transition>
		</div>

		<table class="table is-hoverable is-fullwidth">
			<thead>
			<tr>
				<th v-if="activeColumns.id">
					#
					<sort :order="sortBy.id" @click="sort('id')"/>
				</th>
				<th v-if="activeColumns.done">
					Done
					<sort :order="sortBy.done" @click="sort('done')"/>
				</th>
				<th v-if="activeColumns.title">
					Name
					<sort :order="sortBy.title" @click="sort('title')"/>
				</th>
				<th v-if="activeColumns.priority">
					Priority
					<sort :order="sortBy.priority" @click="sort('priority')"/>
				</th>
				<th v-if="activeColumns.labels">
					Labels
				</th>
				<th v-if="activeColumns.assignees">
					Assignees
				</th>
				<th v-if="activeColumns.dueDate">
					Due&nbsp;Date
					<sort :order="sortBy.due_date_unix" @click="sort('due_date_unix')"/>
				</th>
				<th v-if="activeColumns.startDate">
					Start&nbsp;Date
					<sort :order="sortBy.start_date_unix" @click="sort('start_date_unix')"/>
				</th>
				<th v-if="activeColumns.endDate">
					End&nbsp;Date
					<sort :order="sortBy.end_date_unix" @click="sort('end_date_unix')"/>
				</th>
				<th v-if="activeColumns.percentDone">
					%&nbsp;Done
					<sort :order="sortBy.percent_done" @click="sort('percent_done')"/>
				</th>
				<th v-if="activeColumns.created">
					Created
					<sort :order="sortBy.created" @click="sort('created')"/>
				</th>
				<th v-if="activeColumns.updated">
					Updated
					<sort :order="sortBy.updated" @click="sort('updated')"/>
				</th>
				<th v-if="activeColumns.createdBy">
					Created&nbsp;By
				</th>
			</tr>
			</thead>
			<tbody>
			<tr v-for="t in tasks" :key="t.id">
				<td v-if="activeColumns.id">
					<router-link :to="{name: 'task.detail', params: { id: t.id }}">{{ t.id }}</router-link>
				</td>
				<td v-if="activeColumns.done">
					<div class="is-done" v-if="t.done">Done</div>
				</td>
				<td v-if="activeColumns.title">
					<router-link :to="{name: 'task.detail', params: { id: t.id }}">{{ t.title }}</router-link>
				</td>
				<td v-if="activeColumns.priority">
					<priority-label :priority="t.priority" :show-all="true"/>
				</td>
				<td v-if="activeColumns.labels">
					<labels :labels="t.labels"/>
				</td>
				<td v-if="activeColumns.assignees">
					<user
							:user="a"
							:avatar-size="27"
							:show-username="false"
							:is-inline="true"
							v-for="(a, i) in t.assignees"
							:key="t.id + 'assignee' + a.id + i"
					/>
				</td>
				<date-table-cell :date="t.dueDate" v-if="activeColumns.dueDate"/>
				<date-table-cell :date="t.startDate" v-if="activeColumns.startDate"/>
				<date-table-cell :date="t.endDate" v-if="activeColumns.endDate"/>
				<td v-if="activeColumns.percentDone">{{ t.percentDone * 100 }}%</td>
				<date-table-cell :date="t.created" v-if="activeColumns.created"/>
				<date-table-cell :date="t.updated" v-if="activeColumns.updated"/>
				<td v-if="activeColumns.createdBy">
					<user
							:user="t.createdBy"
							:show-username="false"
							:avatar-size="27"/>
				</td>
			</tr>
			</tbody>
		</table>

		<nav class="pagination is-centered" role="navigation" aria-label="pagination" v-if="taskCollectionService.totalPages > 1">
			<router-link class="pagination-previous" :to="getRouteForPagination(currentPage - 1, 'table')" tag="button" :disabled="currentPage === 1">Previous</router-link>
			<router-link class="pagination-next" :to="getRouteForPagination(currentPage + 1, 'table')" tag="button" :disabled="currentPage === taskCollectionService.totalPages">Next page</router-link>
			<ul class="pagination-list">
				<template v-for="(p, i) in pages">
					<li :key="'page'+i" v-if="p.isEllipsis"><span class="pagination-ellipsis">&hellip;</span></li>
					<li :key="'page'+i" v-else>
						<router-link :to="getRouteForPagination(p.number, 'table')" :class="{'is-current': p.number === currentPage}" class="pagination-link" :aria-label="'Goto page ' + p.number">{{ p.number }}</router-link>
					</li>
				</template>
			</ul>
		</nav>

		<!-- This router view is used to show the task popup while keeping the table view itself -->
		<transition name="modal">
			<router-view/>
		</transition>

	</div>
</template>

<script>
	import taskList from '../../tasks/helpers/taskList'
	import User from '../../global/user'
	import PriorityLabel from '../../tasks/reusable/priorityLabel'
	import Labels from '../../tasks/reusable/labels'
	import DateTableCell from '../../tasks/reusable/date-table-cell'
	import Fancycheckbox from '../../global/fancycheckbox'
	import Sort from '../../tasks/reusable/sort'
	import {saveListView} from '../../../helpers/saveListView'
	import Filters from '../reusable/filters'

	export default {
		name: 'Table',
		components: {
			Filters,
			Sort,
			Fancycheckbox,
			DateTableCell,
			Labels,
			PriorityLabel,
			User,
		},
		mixins: [
			taskList,
		],
		data() {
			return {
				showActiveColumnsFilter: false,
				activeColumns: {
					id: true,
					done: true,
					title: true,
					priority: false,
					labels: true,
					assignees: true,
					dueDate: true,
					startDate: false,
					endDate: false,
					percentDone: false,
					created: false,
					updated: false,
					createdBy: false,
				},
				sortBy: {
					id: 'desc',
				},
			}
		},
		created() {
			const savedShowColumns = localStorage.getItem('tableViewColumns')
			if (savedShowColumns !== null) {
				this.$set(this, 'activeColumns', JSON.parse(savedShowColumns))
			}
			const savedSortBy = localStorage.getItem('tableViewSortBy')
			if (savedSortBy !== null) {
				this.$set(this, 'sortBy', JSON.parse(savedSortBy))
			}

			this.$set(this.params, 'filter_by', [])
			this.$set(this.params, 'filter_value', [])
			this.$set(this.params, 'filter_comparator', [])

			this.initTasks(1)

			// Save the current list view to local storage
			// We use local storage and not vuex here to make it persistent across reloads.
			saveListView(this.$route.params.listId, this.$route.name)
		},
		methods: {
			initTasks(page, search = '') {
				const params = this.params
				params.sort_by = []
				params.order_by = []
				Object.keys(this.sortBy).map(s => {
					params.sort_by.push(s)
					params.order_by.push(this.sortBy[s])
				})
				this.loadTasks(page, search, params)
			},
			sort(property) {
				const order = this.sortBy[property]
				if (typeof order === 'undefined' || order === 'none') {
					this.$set(this.sortBy, property, 'desc')
				} else if (order === 'desc') {
					this.$set(this.sortBy, property, 'asc')
				} else {
					this.$delete(this.sortBy, property)
				}
				this.initTasks(this.currentPage, this.searchTerm)
				// Save the order to be able to retrieve them later
				localStorage.setItem('tableViewSortBy', JSON.stringify(this.sortBy))
			},
			saveTaskColumns() {
				localStorage.setItem('tableViewColumns', JSON.stringify(this.activeColumns))
			},
		},
	}
</script>
