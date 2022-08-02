<template>
	<ListWrapper class="list-gantt" :list-id="props.listId" viewName="gantt">
		<template #header>
			<card>
				<div class="gantt-options">
					<div class="range-picker">
						<div class="field">
							<label class="label" for="precision">{{ $t('list.gantt.size') }}</label>
							<div class="control">
								<div class="select">
									<select id="precision" v-model="precision">
										<option value="day">{{ $t('list.gantt.day') }}</option>
										<option value="month">{{ $t('list.gantt.month') }}</option>
									</select>
								</div>
							</div>
						</div>
						<div class="field">
							<label class="label" for="range">{{ $t('list.gantt.range') }}</label>
							<div class="control">
								<flat-pickr
									:config="flatPickerConfig"
									class="input"
									id="range"
									:placeholder="$t('list.gantt.range')"
									v-model="range"
								/>
							</div>
						</div>
					</div>
					<fancycheckbox class="is-block" v-model="showTasksWithoutDates">
						{{ $t('list.gantt.showTasksWithoutDates') }}
					</fancycheckbox>
				</div>
			</card>
		</template>

		<template #default>
			<div class="gantt-chart-container">
				<card :padding="false" class="has-overflow">

					<gantt-chart
						:date-from="dateFrom"
						:date-to="dateTo"
						:precision="precision"
						:list-id="props.listId"
						:show-tasks-without-dates="showTasksWithoutDates"
					/>

				</card>
			</div>
		</template>
	</ListWrapper>
</template>

<script setup lang="ts">
import {ref, computed} from 'vue'
import flatPickr from 'vue-flatpickr-component'
import {useI18n} from 'vue-i18n'

import {useAuthStore} from '@/stores/auth'

import ListWrapper from './ListWrapper.vue'
import GanttChart from '@/components/tasks/gantt-chart.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import {format} from 'date-fns'

const props = defineProps({
	listId: {
		type: Number,
		required: true,
	},
})

const showTasksWithoutDates = ref(false)
const precision = ref('day')

const now = new Date()
const defaultFrom = format(new Date((new Date()).setDate(now.getDate() - 15)), 'yyyy-LL-dd')
const defaultTo = format(new Date((new Date()).setDate(now.getDate() + 55)), 'yyyy-LL-dd')
const range = ref(`${defaultFrom} to ${defaultTo}`)

// TODO: only update once both dates are available (maybe use a watcher + refs instead?)
const dateFrom = computed(() => range.value?.split(' to ')[0] ?? defaultFrom)
const dateTo = computed(() => range.value?.split(' to ')[1] ?? defaultTo)

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatShort'),
	altInput: true,
	dateFormat: 'Y-m-d',
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

	.range-picker {
		display: flex;
		margin-bottom: 1rem;
		width: 50%;

		@media screen and (max-width: $tablet) {
			flex-direction: column;
			width: 100%;
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
				padding-left: .4rem;
			}
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