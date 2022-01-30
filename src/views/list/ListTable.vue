<template>
	<ListWrapper class="list-table" :list-id="listId" viewName="table">
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
							<fancycheckbox v-model="activeColumns.id">#</fancycheckbox>
							<fancycheckbox v-model="activeColumns.done">
								{{ $t('task.attributes.done') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.title">
								{{ $t('task.attributes.title') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.priority">
								{{ $t('task.attributes.priority') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.labels">
								{{ $t('task.attributes.labels') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.assignees">
								{{ $t('task.attributes.assignees') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.dueDate">
								{{ $t('task.attributes.dueDate') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.startDate">
								{{ $t('task.attributes.startDate') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.endDate">
								{{ $t('task.attributes.endDate') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.percentDone">
								{{ $t('task.attributes.percentDone') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.created">
								{{ $t('task.attributes.created') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.updated">
								{{ $t('task.attributes.updated') }}
							</fancycheckbox>
							<fancycheckbox v-model="activeColumns.createdBy">
								{{ $t('task.attributes.createdBy') }}
							</fancycheckbox>
						</card>
					</template>
				</popup>
				<filter-popup v-model="params" />
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

<script setup lang="ts">
import { toRef, computed, Ref } from 'vue'
import { useRouter } from 'vue-router'

import { useStorage } from '@vueuse/core'

import ListWrapper from './ListWrapper.vue'
import Done from '@/components/misc/Done.vue'
import User from '@/components/misc/user.vue'
import PriorityLabel from '@/components/tasks/partials/priorityLabel.vue'
import Labels from '@/components/tasks/partials/labels.vue'
import DateTableCell from '@/components/tasks/partials/date-table-cell.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import Sort from '@/components/tasks/partials/sort.vue'
import FilterPopup from '@/components/list/partials/filter-popup.vue'
import Pagination from '@/components/misc/pagination.vue'
import Popup from '@/components/misc/popup.vue'

import { useTaskList } from '@/composables/taskList'
import TaskModel from '@/models/task'

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

const props = defineProps({
	listId: {
		type: Number,
		required: true,
	},
})

type Order = 'asc' | 'desc' | 'none'

interface SortBy {
	id : Order
	done? : Order 
	title? : Order 
	priority? : Order 
	due_date? : Order 
	start_date? : Order 
	end_date? : Order 
	percent_done? : Order 
	created? : Order 
	updated? : Order 
}

const SORT_BY_DEFAULT : SortBy = {
	id: 'desc',
}

const activeColumns = useStorage('tableViewColumns', { ...ACTIVE_COLUMNS_DEFAULT })
const sortBy = useStorage<SortBy>('tableViewSortBy', { ...SORT_BY_DEFAULT })

const taskList = useTaskList(toRef(props, 'listId'))

const {
	loading,
	params,
	totalPages,
	currentPage,
} = taskList
const tasks : Ref<TaskModel[]> = taskList.tasks

Object.assign(params.value, {
	filter_by: [],
	filter_value: [],
	filter_comparator: [],
})

// FIXME: by doing this we can have multiple sort orders
function sort(property : keyof SortBy) {
	const order = sortBy.value[property]
	if (typeof order === 'undefined' || order === 'none') {
		sortBy.value[property] = 'desc'
	} else if (order === 'desc') {
		sortBy.value[property] = 'asc'
	} else {
		delete sortBy.value[property]
	}
}

const router = useRouter()
const taskDetailRoutes = computed(() => Object.fromEntries(
	tasks.value.map(({id}) => ([
		id,
		{
			name: 'task.detail',
			params: { id },
			state: { backdropView: router.currentRoute.value.fullPath },
		},
	])),
))
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