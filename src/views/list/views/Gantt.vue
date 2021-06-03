<template>
	<div class="gantt-chart-container">
		<card :padding="false" class="has-overflow">
		<div class="gantt-options p-4">
			<fancycheckbox class="is-block" v-model="showTaskswithoutDates">
				Show tasks which don't have dates set
			</fancycheckbox>
			<div class="range-picker">
				<div class="field">
					<label class="label" for="dayWidth">Size</label>
					<div class="control">
						<div class="select">
							<select id="dayWidth" v-model.number="dayWidth">
								<option value="35">Default</option>
								<option value="10">Month</option>
								<option value="80">Day</option>
							</select>
						</div>
					</div>
				</div>
				<div class="field">
					<label class="label" for="fromDate">From</label>
					<div class="control">
						<flat-pickr
							:config="flatPickerConfig"
							class="input"
							id="fromDate"
							placeholder="From"
							v-model="dateFrom"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="toDate">To</label>
					<div class="control">
						<flat-pickr
							:config="flatPickerConfig"
							class="input"
							id="toDate"
							placeholder="To"
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
import {mapState} from 'vuex'

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
	computed: mapState({
		flatPickerConfig: state => ({
			altFormat: 'j M Y',
			altInput: true,
			dateFormat: 'Y-m-d',
			enableTime: false,
			locale: {
				firstDayOfWeek: state.auth.settings.weekStart,
			},
		})
	}),
	beforeMount() {
		this.dateFrom = new Date((new Date()).setDate((new Date()).getDate() - 15))
		this.dateTo = new Date((new Date()).setDate((new Date()).getDate() + 30))
	},
}
</script>