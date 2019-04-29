<template>
	<div>
		<div class="gantt-options">
			<div class="fancycheckbox is-block">
				<input id="showTaskswithoutDates" type="checkbox" style="display: none;" v-model="showTaskswithoutDates">
				<label for="showTaskswithoutDates" class="check">
					<svg width="18px" height="18px" viewBox="0 0 18 18">
						<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
						<polyline points="1 9 7 14 15 4"></polyline>
					</svg>
					<span>
						Show tasks which don't have dates set
					</span>
				</label>
			</div>
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
			:list="list"
			:show-taskswithout-dates="showTaskswithoutDates"
			:date-from="dateFrom"
			:date-to="dateTo"
			:day-width="dayWidth"
		/>
	</div>
</template>

<script>
	import GanttChart from './gantt-component'
	import flatPickr from 'vue-flatpickr-component'
	import ListModel from '../../models/list'

	export default {
		name: 'Gantt',
		components: {
			flatPickr,
			GanttChart
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
		props: {
			list: {
				type: ListModel,
				required: true,
			}
		},
	}
</script>