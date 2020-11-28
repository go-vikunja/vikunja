<template>
	<div class="datepicker" :class="{'disabled': disabled}">
		<a @click.stop="toggleDatePopup" class="show">
			<template v-if="date === null">
				{{ chooseDateLabel }}
			</template>
			<template v-else>
				{{ formatDateShort(date) }}
			</template>
		</a>

		<transition name="fade">
			<div v-if="show" class="datepicker-popup" ref="datepickerPopup">

				<a @click.stop="() => setDate('today')" v-if="(new Date()).getHours() < 21">
					<span class="icon">
						<icon :icon="['far', 'calendar-alt']"/>
					</span>
					<span class="text">
						<span>
							Today
						</span>
						<span class="weekday">
							{{ getWeekdayFromStringInterval('today') }}
						</span>
					</span>
				</a>
				<a @click.stop="() => setDate('tomorrow')">
					<span class="icon">
						<icon :icon="['far', 'sun']"/>
					</span>
					<span class="text">
						<span>
							Tomorrow
						</span>
						<span class="weekday">
							{{ getWeekdayFromStringInterval('tomorrow') }}
						</span>
					</span>
				</a>
				<a @click.stop="() => setDate('nextMonday')">
					<span class="icon">
						<icon icon="coffee"/>
					</span>
					<span class="text">
						<span>
							Next Monday
						</span>
						<span class="weekday">
							{{ getWeekdayFromStringInterval('nextMonday') }}
						</span>
					</span>
				</a>
				<a @click.stop="() => setDate('thisWeekend')">
					<span class="icon">
						<icon icon="cocktail"/>
					</span>
					<span class="text">
						<span>
							This Weekend
						</span>
						<span class="weekday">
							{{ getWeekdayFromStringInterval('thisWeekend') }}
						</span>
					</span>
				</a>
				<a @click.stop="() => setDate('laterThisWeek')">
					<span class="icon">
						<icon icon="chess-knight"/>
					</span>
					<span class="text">
						<span>
							Later This Week
						</span>
						<span class="weekday">
							{{ getWeekdayFromStringInterval('laterThisWeek') }}
						</span>
					</span>
				</a>
				<a @click.stop="() => setDate('nextWeek')">
					<span class="icon">
						<icon icon="forward"/>
					</span>
					<span class="text">
						<span>
							Next Week
						</span>
						<span class="weekday">
							{{ getWeekdayFromStringInterval('nextWeek') }}
						</span>
					</span>
				</a>

				<flat-pickr
					:config="flatPickerConfig"
					class="input"
					v-model="flatPickrDate"
				/>

				<a
					class="button is-outlined is-primary has-no-shadow is-fullwidth"
					@click="close"
				>
					Confirm
				</a>
			</div>
		</transition>
	</div>
</template>

<script>
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

import {calculateDayInterval} from '@/helpers/time/calculateDayInterval'
import {format} from 'date-fns'
import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'

export default {
	name: 'datepicker',
	data() {
		return {
			date: null,
			show: false,
			changed: false,

			flatPickerConfig: {
				altFormat: 'j M Y H:i',
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				time_24hr: true,
				inline: true,
			},
			// Since flatpickr dates are strings, we need to convert them to native date objects.
			// To make that work, we need a separate variable since flatpickr does not have a change event.
			flatPickrDate: null,
		}
	},
	components: {
		flatPickr,
	},
	props: {
		value: {
			validator: prop => prop instanceof Date || prop === null || typeof prop === 'string'
		},
		chooseDateLabel: {
			type: String,
			default: 'Choose a date'
		},
		disabled: {
			type: Boolean,
			default: false,
		}
	},
	mounted() {
		this.date = this.value
		document.addEventListener('click', this.hideDatePopup)
	},
	beforeDestroy() {
		document.removeEventListener('click', this.hideDatePopup)
	},
	watch: {
		value(newVal) {
			if(newVal === null) {
				this.date = null
				return
			}
			this.date = new Date(newVal)
		},
		flatPickrDate(newVal) {
			this.date = new Date(newVal)
			this.updateData()
		},
	},
	methods: {
		updateData() {
			this.changed = true
			this.$emit('input', this.date)
			this.$emit('change', this.date)
		},
		toggleDatePopup() {
			if(this.disabled) {
				return
			}

			this.show = !this.show
		},
		hideDatePopup(e) {
			if (this.show) {

				// We walk up the tree to see if any parent of the clicked element is the datepicker element.
				// If it is not, we hide the popup. We're doing all this hassle to prevent the popup from closing when
				// clicking an element of flatpickr.
				let parent = e.target.parentElement
				while (parent !== this.$refs.datepickerPopup) {
					if (parent.parentElement === null) {
						parent = null
						break
					}

					parent = parent.parentElement
				}

				if (parent === this.$refs.datepickerPopup) {
					return
				}

				this.close()
			}
		},
		close() {
			this.show = false
			this.$emit('close', this.changed)
			if(this.changed) {
				this.changed = false
				this.$emit('close-on-change', this.changed)
			}
		},
		setDate(date) {
			if (this.date === null) {
				this.date = new Date()
			}

			const interval = calculateDayInterval(date)
			const newDate = new Date()
			newDate.setDate(newDate.getDate() + interval)
			newDate.setHours(calculateNearestHours(newDate))
			newDate.setMinutes(0)
			newDate.setSeconds(0)
			this.date = newDate
			this.flatPickrDate = newDate
			this.updateData()
		},
		getDayIntervalFromString(date) {
			return calculateDayInterval(date)
		},
		getWeekdayFromStringInterval(date) {
			const interval = calculateDayInterval(date)
			const newDate = new Date()
			newDate.setDate(newDate.getDate() + interval)
			return format(newDate, 'E')
		},
	},
}
</script>
