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
							{{ $t('input.datepicker.today') }}
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
							{{ $t('input.datepicker.tomorrow') }}
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
							{{ $t('input.datepicker.nextMonday') }}
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
							{{ $t('input.datepicker.thisWeekend') }}
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
							{{ $t('input.datepicker.laterThisWeek') }}
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
							{{ $t('input.datepicker.nextWeek') }}
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

				<x-button
					class="is-fullwidth"
					:shadow="false"
					@click="close"
				>
					{{ $t('misc.confirm') }}
				</x-button>
			</div>
		</transition>
	</div>
</template>

<script>
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

import {format} from 'date-fns'
import {calculateDayInterval} from '@/helpers/time/calculateDayInterval'
import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {createDateFromString} from '@/helpers/time/createDateFromString'

export default {
	name: 'datepicker',
	data() {
		return {
			date: null,
			show: false,
			changed: false,

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
			validator: prop => prop instanceof Date || prop === null || typeof prop === 'string',
		},
		chooseDateLabel: {
			type: String,
			default() {
				return this.$t('input.datepicker.chooseDate')
			},
		},
		disabled: {
			type: Boolean,
			default: false,
		},
	},
	mounted() {
		this.setDateValue(this.value)
		document.addEventListener('click', this.hideDatePopup)
	},
	beforeDestroy() {
		document.removeEventListener('click', this.hideDatePopup)
	},
	watch: {
		value(newVal) {
			this.setDateValue(newVal)
		},
		flatPickrDate(newVal) {
			this.date = createDateFromString(newVal)
			this.updateData()
		},
	},
	computed: {
		flatPickerConfig() {
			return {
				altFormat: this.$t('date.altFormatLong'),
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				time_24hr: true,
				inline: true,
				locale: {
					firstDayOfWeek: this.$store.state.auth.settings.weekStart,
				},
			}
		},
	},
	methods: {
		setDateValue(newVal) {
			if(newVal === null) {
				this.date = null
				return
			}
			this.date = createDateFromString(newVal)
		},
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
				closeWhenClickedOutside(e, this.$refs.datepickerPopup, this.close)
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
