<template>
	<div class="gantt-chart-container">
		<card :padding="false" class="has-overflow">
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
		<gantt-chart
			:date-from="dateFrom"
			:date-to="dateTo"
			:day-width="dayWidth"
			:list-id="Number($route.params.listId)"
			:show-taskswithout-dates="showTaskswithoutDates"
		/>

		<!-- This router view is used to show the task popup while keeping the gantt chart itself -->
		<transition name="modal">
			<router-view/>
		</transition>

		</card>
	</div>
</template>

<script>
import GanttChart from '../../../components/tasks/gantt-component'
import flatPickr from 'vue-flatpickr-component'
import Fancycheckbox from '../../../components/input/fancycheckbox'
import {saveListView} from '@/helpers/saveListView'

export default {
	name: 'Gantt',
	components: {
		Fancycheckbox,
		flatPickr,
		GanttChart,
	},
	created() {
		// Save the current list view to local storage
		// We use local storage and not vuex here to make it persistent across reloads.
		saveListView(this.$route.params.listId, this.$route.name)
	},
	data() {
		return {
			showTaskswithoutDates: false,
			dayWidth: 35,
			dateFrom: null,
			dateTo: null,
		}
	},
	computed: {
		flatPickerConfig() {
			return {
				altFormat: this.$t('date.altFormatShort'),
				altInput: true,
				dateFormat: 'Y-m-d',
				enableTime: false,
				locale: {
					firstDayOfWeek: this.$store.state.auth.settings.weekStart,
				},
			}
		},
	},
	beforeMount() {
		this.dateFrom = new Date((new Date()).setDate((new Date()).getDate() - 15))
		this.dateTo = new Date((new Date()).setDate((new Date()).getDate() + 30))
	},
}
</script>