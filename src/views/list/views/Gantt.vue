<template>
	<div class="gantt-chart-container">
		<div class="gantt-options">
			<fancycheckbox v-model="showTaskswithoutDates" class="is-block">
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
							class="input"
							v-model="dateFrom"
							:config="flatPickerConfig"
							id="fromDate"
							placeholder="From"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="toDate">To</label>
					<div class="control">
						<flat-pickr
							class="input"
							v-model="dateTo"
							:config="flatPickerConfig"
							id="toDate"
							placeholder="To"
						/>
					</div>
				</div>
			</div>
		</div>
		<gantt-chart
			:list-id="Number($route.params.listId)"
			:show-taskswithout-dates="showTaskswithoutDates"
			:date-from="dateFrom"
			:date-to="dateTo"
			:day-width="dayWidth"
		/>

		<!-- This router view is used to show the task popup while keeping the gantt chart itself -->
		<transition name="modal">
			<router-view/>
		</transition>

	</div>
</template>

<script>
	import GanttChart from '../../../components/tasks/gantt-component'
	import flatPickr from 'vue-flatpickr-component'
	import Fancycheckbox from '../../../components/input/fancycheckbox'
	import {saveListView} from '../../../helpers/saveListView'

	export default {
		name: 'Gantt',
		components: {
			Fancycheckbox,
			flatPickr,
			GanttChart
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
				flatPickerConfig:{
					altFormat: 'j M Y',
					altInput: true,
					dateFormat: 'Y-m-d',
					enableTime: false,
				},
			}
		},
		beforeMount() {
			this.dateFrom = new Date((new Date()).setDate((new Date()).getDate() - 15))
			this.dateTo = new Date((new Date()).setDate((new Date()).getDate() + 30))
		},
	}
</script>