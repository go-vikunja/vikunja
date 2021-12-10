<template>
	<ListWrapper class="list-gantt" :list-id="props.listId" viewName="gantt">
		<template #header>
		<div class="gantt-options p-4">
			<fancycheckbox class="is-block" v-model="showTaskswithoutDates">
				{{ $t('list.gantt.showTasksWithoutDates') }}
			</fancycheckbox>
			<div class="range-picker">
				<div class="field">
					<label class="label" for="dayWidth">{{ $t('list.gantt.size') }}</label>
					<div class="control">
						<div class="select">
							<select id="dayWidth" v-model.number="dayWidth">
								<option value="35">{{ $t('list.gantt.default') }}</option>
								<option value="10">{{ $t('list.gantt.month') }}</option>
								<option value="80">{{ $t('list.gantt.day') }}</option>
							</select>
						</div>
					</div>
				</div>
				<div class="field">
					<label class="label" for="fromDate">{{ $t('list.gantt.from') }}</label>
					<div class="control">
						<flat-pickr
							:config="flatPickerConfig"
							class="input"
							id="fromDate"
							:placeholder="$t('list.gantt.from')"
							v-model="dateFrom"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="toDate">{{ $t('list.gantt.to') }}</label>
					<div class="control">
						<flat-pickr
							:config="flatPickerConfig"
							class="input"
							id="toDate"
							:placeholder="$t('list.gantt.to')"
							v-model="dateTo"
						/>
					</div>
				</div>
			</div>
		</div>
		</template>
		
		<template #default>
		<div class="gantt-chart-container">
		<card :padding="false" class="has-overflow">

		<gantt-chart
			:date-from="dateFrom"
			:date-to="dateTo"
			:day-width="dayWidth"
			:list-id="props.listId"
			:show-taskswithout-dates="showTaskswithoutDates"
		/>

		</card>
		</div>
		</template>
	</ListWrapper>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import flatPickr from 'vue-flatpickr-component'

import { useI18n } from 'vue-i18n'
import { useStore } from 'vuex'

import ListWrapper from './ListWrapper.vue'
import GanttChart from '@/components/tasks/gantt-component.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'

const props = defineProps({
	listId: {
		type: Number,
		required: true,
	},
})

const DEFAULT_DAY_COUNT = 35 

const showTaskswithoutDates = ref(false)
const dayWidth = ref(DEFAULT_DAY_COUNT)

const now = ref(new Date())
const dateFrom = ref(new Date((new Date()).setDate(now.value.getDate() - 15)))
const dateTo = ref(new Date((new Date()).setDate(now.value.getDate() + 30)))

const {t} = useI18n()
const store = useStore()
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatShort'),
	altInput: true,
	dateFormat: 'Y-m-d',
	enableTime: false,
	locale: {
		firstDayOfWeek: store.state.auth.settings.weekStart,
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
</style>