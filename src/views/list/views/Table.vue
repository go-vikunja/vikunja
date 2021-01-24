<template>
	<div :class="{'is-loading': taskCollectionService.loading}" class="table-view loader-container">
		<div class="filter-container">
			<div class="items">
				<x-button
					@click.prevent.stop="() => {showActiveColumnsFilter = !showActiveColumnsFilter; showTaskFilter = false}"
					icon="th"
					type="secondary"
				>
					Columns
				</x-button>
				<x-button
					@click.prevent.stop="() => {showTaskFilter = !showTaskFilter; showActiveColumnsFilter = false}"
					icon="filter"
					type="secondary"
				>
					Filters
				</x-button>
			</div>
			<transition name="fade">
				<card v-if="showActiveColumnsFilter">
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
				</card>
			</transition>
			<filter-popup
				@change="loadTasks(1)"
				:visible="showTaskFilter"
				v-model="params"
			/>
		</div>

		<card :padding="false" :has-content="false">
			<table class="table is-hoverable is-fullwidth mb-0">
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
						<sort :order="sortBy.due_date" @click="sort('due_date')"/>
					</th>
					<th v-if="activeColumns.startDate">
						Start&nbsp;Date
						<sort :order="sortBy.start_date" @click="sort('start_date')"/>
					</th>
					<th v-if="activeColumns.endDate">
						End&nbsp;Date
						<sort :order="sortBy.end_date" @click="sort('end_date')"/>
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
				<tr :key="t.id" v-for="t in tasks">
					<td v-if="activeColumns.id">
						<router-link :to="{name: 'task.detail', params: { id: t.id }}">
							<template v-if="t.identifier === ''">
								#{{ t.index }}
							</template>
							<template v-else>
								{{ t.identifier }}
							</template>
						</router-link>
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
							:avatar-size="27"
							:is-inline="true"
							:key="t.id + 'assignee' + a.id + i"
							:show-username="false"
							:user="a"
							v-for="(a, i) in t.assignees"
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
							:avatar-size="27"
							:show-username="false"
							:user="t.createdBy"/>
					</td>
				</tr>
				</tbody>
			</table>

			<nav
				aria-label="pagination"
				class="pagination is-centered p-4"
				role="navigation"
				v-if="taskCollectionService.totalPages > 1">
				<router-link
					:disabled="currentPage === 1"
					:to="getRouteForPagination(currentPage - 1, 'table')"
					class="pagination-previous"
					tag="button">
					Previous
				</router-link>
				<router-link
					:disabled="currentPage === taskCollectionService.totalPages"
					:to="getRouteForPagination(currentPage + 1, 'table')"
					class="pagination-next"
					tag="button">
					Next page
				</router-link>
				<ul class="pagination-list">
					<template v-for="(p, i) in pages">
						<li :key="'page'+i" v-if="p.isEllipsis"><span class="pagination-ellipsis">&hellip;</span></li>
						<li :key="'page'+i" v-else>
							<router-link
								:aria-label="'Goto page ' + p.number"
								:class="{'is-current': p.number === currentPage}"
								:to="getRouteForPagination(p.number, 'table')"
								class="pagination-link">
								{{ p.number }}
							</router-link>
						</li>
					</template>
				</ul>
			</nav>
		</card>

		<!-- This router view is used to show the task popup while keeping the table view itself -->
		<transition name="modal">
			<router-view/>
		</transition>

	</div>
</template>

<script>
import taskList from '../../../components/tasks/mixins/taskList'
import User from '../../../components/misc/user'
import PriorityLabel from '../../../components/tasks/partials/priorityLabel'
import Labels from '../../../components/tasks/partials/labels'
import DateTableCell from '../../../components/tasks/partials/date-table-cell'
import Fancycheckbox from '../../../components/input/fancycheckbox'
import Sort from '../../../components/tasks/partials/sort'
import {saveListView} from '@/helpers/saveListView'
import FilterPopup from '@/components/list/partials/filter-popup'

export default {
	name: 'Table',
	components: {
		FilterPopup,
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
