<template>
	<ListWrapper class="list-table">
		<template #header>
			<div class="filter-container">
			<div class="items">
				<popup>
					<template #trigger="{toggle}">
						<x-button
							@click.prevent.stop="toggle()"
							icon="th"
							variant="secondary"
						>
							{{ $t('list.table.columns') }}
						</x-button>
					</template>
					<template #content="{isOpen}">
						<card class="columns-filter" :class="{'is-open': isOpen}">
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.id">#</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.done">
								{{ $t('task.attributes.done') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.title">
								{{ $t('task.attributes.title') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.priority">
								{{ $t('task.attributes.priority') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.labels">
								{{ $t('task.attributes.labels') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.assignees">
								{{ $t('task.attributes.assignees') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.dueDate">
								{{ $t('task.attributes.dueDate') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.startDate">
								{{ $t('task.attributes.startDate') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.endDate">
								{{ $t('task.attributes.endDate') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.percentDone">
								{{ $t('task.attributes.percentDone') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.created">
								{{ $t('task.attributes.created') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.updated">
								{{ $t('task.attributes.updated') }}
							</fancycheckbox>
							<fancycheckbox @change="saveTaskColumns" v-model="activeColumns.createdBy">
								{{ $t('task.attributes.createdBy') }}
							</fancycheckbox>
						</card>
					</template>
				</popup>
				<filter-popup
					v-model="params"
					@update:modelValue="loadTasks()"
				/>
			</div>
			</div>
		</template>

		<template #default>
			<div :class="{'is-loading': loading}" class="loader-container">
			<card :padding="false" :has-content="false">
			<div class="has-horizontal-overflow">
				<table class="table has-actions is-hoverable is-fullwidth mb-0">
					<thead>
					<tr>
						<th v-if="activeColumns.id">
							#
							<Sort :order="sortBy.id" @click="sort('id')"/>
						</th>
						<th v-if="activeColumns.done">
							{{ $t('task.attributes.done') }}
							<Sort :order="sortBy.done" @click="sort('done')"/>
						</th>
						<th v-if="activeColumns.title">
							{{ $t('task.attributes.title') }}
							<Sort :order="sortBy.title" @click="sort('title')"/>
						</th>
						<th v-if="activeColumns.priority">
							{{ $t('task.attributes.priority') }}
							<Sort :order="sortBy.priority" @click="sort('priority')"/>
						</th>
						<th v-if="activeColumns.labels">
							{{ $t('task.attributes.labels') }}
						</th>
						<th v-if="activeColumns.assignees">
							{{ $t('task.attributes.assignees') }}
						</th>
						<th v-if="activeColumns.dueDate">
							{{ $t('task.attributes.dueDate') }}
							<Sort :order="sortBy.due_date" @click="sort('due_date')"/>
						</th>
						<th v-if="activeColumns.startDate">
							{{ $t('task.attributes.startDate') }}
							<Sort :order="sortBy.start_date" @click="sort('start_date')"/>
						</th>
						<th v-if="activeColumns.endDate">
							{{ $t('task.attributes.endDate') }}
							<Sort :order="sortBy.end_date" @click="sort('end_date')"/>
						</th>
						<th v-if="activeColumns.percentDone">
							{{ $t('task.attributes.percentDone') }}
							<Sort :order="sortBy.percent_done" @click="sort('percent_done')"/>
						</th>
						<th v-if="activeColumns.created">
							{{ $t('task.attributes.created') }}
							<Sort :order="sortBy.created" @click="sort('created')"/>
						</th>
						<th v-if="activeColumns.updated">
							{{ $t('task.attributes.updated') }}
							<Sort :order="sortBy.updated" @click="sort('updated')"/>
						</th>
						<th v-if="activeColumns.createdBy">
							{{ $t('task.attributes.createdBy') }}
						</th>
					</tr>
					</thead>
					<tbody>
					<tr :key="t.id" v-for="t in tasks">
						<td v-if="activeColumns.id">
							<router-link :to="taskDetailRoutes[t.id]">
								<template v-if="t.identifier === ''">
									#{{ t.index }}
								</template>
								<template v-else>
									{{ t.identifier }}
								</template>
							</router-link>
						</td>
						<td v-if="activeColumns.done">
							<Done :is-done="t.done" variant="small" />
						</td>
						<td v-if="activeColumns.title">
							<router-link :to="taskDetailRoutes[t.id]">{{ t.title }}</router-link>
						</td>
						<td v-if="activeColumns.priority">
							<priority-label :priority="t.priority" :done="t.done" :show-all="true"/>
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
			</div>

			<Pagination 
				:total-pages="totalPages"
				:current-page="currentPage"
			/>
			</card>
			</div>
		</template>
	</ListWrapper>
</template>

<script setup>
import { ref, reactive, computed, toRaw } from 'vue'
import { useRouter } from 'vue-router'

import ListWrapper from './ListWrapper'
import Done from '@/components/misc/Done.vue'
import User from '@/components/misc/user'
import PriorityLabel from '@/components/tasks/partials/priorityLabel'
import Labels from '@/components/tasks/partials/labels'
import DateTableCell from '@/components/tasks/partials/date-table-cell'
import Fancycheckbox from '@/components/input/fancycheckbox'
import Sort from '@/components/tasks/partials/sort'
import FilterPopup from '@/components/list/partials/filter-popup.vue'
import Pagination from '@/components/misc/pagination.vue'
import Popup from '@/components/misc/popup'

import { useTaskList } from '@/composables/taskList'

const ACTIVE_COLUMNS_DEFAULT = {
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
}

const SORT_BY_DEFAULT = {
	id: 'desc',
}

function useSavedView(activeColumns, sortBy) {
	const savedShowColumns = localStorage.getItem('tableViewColumns')
	if (savedShowColumns !== null) {
		Object.assign(activeColumns, JSON.parse(savedShowColumns))
	}

	const savedSortBy = localStorage.getItem('tableViewSortBy')
	if (savedSortBy !== null) {
		sortBy.value = JSON.parse(savedSortBy)
	}
}

const activeColumns = reactive({ ...ACTIVE_COLUMNS_DEFAULT })
const sortBy = ref({ ...SORT_BY_DEFAULT })

useSavedView(activeColumns, sortBy)

function beforeLoad(params) {
	// This makes sure an id sort order is always sorted last.
	// When tasks would be sorted first by id and then by whatever else was specified, the id sort takes
	// precedence over everything else, making any other sort columns pretty useless.
	let hasIdFilter = false
	const sortKeys = Object.keys(sortBy.value)
	for (const s of sortKeys) {
		if (s === 'id') {
			sortKeys.splice(s, 1)
			hasIdFilter = true
			break
		}
	}
	if (hasIdFilter) {
		sortKeys.push('id')
	}
	params.value.sort_by = sortKeys
	params.value.order_by = sortKeys.map(s => sortBy.value[s])
}

const {
	tasks,
	loading,
	params,
	loadTasks,
	totalPages,
	currentPage,
	searchTerm,
	initTaskList,
} = useTaskList(beforeLoad)

Object.assign(params.value, {
	filter_by: [],
	filter_value: [],
	filter_comparator: [],
})

const router = useRouter()

const taskDetailRoutes = computed(() => Object.fromEntries(
	tasks.value.map(({id}) => ([
		id,
		{
			name: 'task.detail',
			params: { id },
			state: { backgroundView: router.currentRoute.value.fullPath },
		},
	])),
))

function sort(property) {
	const order = sortBy.value[property]
	if (typeof order === 'undefined' || order === 'none') {
		sortBy.value[property] = 'desc'
	} else if (order === 'desc') {
		sortBy.value[property] = 'asc'
	} else {
		delete sortBy.value[property]
	}
	beforeLoad(currentPage.value, searchTerm.value)
	// Save the order to be able to retrieve them later
	localStorage.setItem('tableViewSortBy', JSON.stringify(sortBy.value))
}

function saveTaskColumns() {
	localStorage.setItem('tableViewColumns', JSON.stringify(toRaw(activeColumns)))
}

initTaskList()
</script>

<style lang="scss" scoped>
.table {
	background: transparent;
	overflow-x: auto;
	overflow-y: hidden;

	th {
		white-space: nowrap;
	}

	.user {
		margin: 0;
	}
}

.columns-filter {
	margin: 0;

	&.is-open {
		margin: 2rem 0 1rem;
	}
}
</style>