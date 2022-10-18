<template>
	<ListWrapper class="list-gantt" :list-id="filters.listId" viewName="gantt">
		<template #header>
			<card>
				<div class="gantt-options">
					<div class="field">
						<label class="label" for="range">{{ $t('list.gantt.range') }}</label>
						<div class="control">
							<Foo
								ref="flatPickerEl"
								:config="flatPickerConfig"
								class="input"
								id="range"
								:placeholder="$t('list.gantt.range')"
								v-model="flatPickerDateRange"
							/>
						</div>
					</div>
					<fancycheckbox class="is-block" v-model="filters.showTasksWithoutDates">
						{{ $t('list.gantt.showTasksWithoutDates') }}
					</fancycheckbox>
				</div>
			</card>
		</template>

		<template #default>
			<div class="gantt-chart-container">
				<card :padding="false" class="has-overflow">
					<gantt-chart
						:list-id="filters.listId"
						:date-from="filters.dateFrom"
						:date-to="filters.dateTo"
						:show-tasks-without-dates="filters.showTasksWithoutDates"
					/>
				</card>
			</div>
		</template>
	</ListWrapper>
</template>

<script setup lang="ts">
import {computed, ref, toRefs} from 'vue'
import Foo from '@/components/misc/flatpickr/Flatpickr.vue'
import type Flatpickr from 'flatpickr'
import {useI18n} from 'vue-i18n'
import type {RouteLocationNormalized} from 'vue-router'

import {useAuthStore} from '@/stores/auth'

import ListWrapper from './ListWrapper.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'

import {createAsyncComponent} from '@/helpers/createAsyncComponent'
import {useGanttFilter} from './helpers/useGanttFilter'

type Options = Flatpickr.Options.Options

const GanttChart = createAsyncComponent(() => import('@/components/tasks/gantt-chart.vue'))

const props = defineProps<{route: RouteLocationNormalized}>()

const {route} = toRefs(props)
const {filters} = useGanttFilter(route)

const flatPickerEl = ref<typeof Foo | null>(null)
const flatPickerDateRange = computed<Date[]>({
	get: () => ([
		new Date(filters.dateFrom),
		new Date(filters.dateTo),
	]),
	set(newVal) {
		const [dateFrom, dateTo] = newVal.map((date) => date?.toISOString())
		
		// only set after whole range has been selected
		if (!dateTo) return

		Object.assign(filters, {dateFrom, dateTo})
	},
})

const initialDateRange = [filters.dateFrom, filters.dateTo]

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const flatPickerConfig = computed<Options>(() => ({
	altFormat: t('date.altFormatShort'),
	altInput: true,
	defaultDate: initialDateRange,
	enableTime: false,
	mode: 'range',
	locale: {
		firstDayOfWeek: authStore.settings.weekStart,
	},
}))
</script>

<style lang="scss">
.gantt-chart-container {
	padding-bottom: 1rem;
}

.gantt-options {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 1rem;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}

	.field {
		margin-bottom: 0;
		width: 33%;

		&:not(:last-child) {
			padding-right: .5rem;
		}

		@media screen and (max-width: $tablet) {
			width: 100%;
			max-width: 100%;
			margin-top: .5rem;
			padding-right: 0 !important;
		}

		&, .input {
			font-size: .8rem;
		}

		.select, .select select {
			height: auto;
			width: 100%;
			font-size: .8rem;
		}


		.label {
			font-size: .9rem;
		}
	}
}

// vue-draggable overwrites
.vdr.active::before {
	display: none;
}

.link-share-view:not(.has-background) .card.gantt-options {
	border: none;
	box-shadow: none;

	.card-content {
		padding: .5rem;
	}
}
</style>