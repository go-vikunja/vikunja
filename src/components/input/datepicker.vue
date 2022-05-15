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
					v-cy="'closeDatepicker'"
				>
					{{ $t('misc.confirm') }}
				</x-button>
			</div>
		</transition>
	</div>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {i18n} from '@/i18n'

import {format} from 'date-fns'
import {calculateDayInterval} from '@/helpers/time/calculateDayInterval'
import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {createDateFromString} from '@/helpers/time/createDateFromString'

export default defineComponent({
	name: 'datepicker',
	data() {
		return {
			date: null,
			show: false,
			changed: false,
		}
	},
	components: {
		flatPickr,
	},
	props: {
		modelValue: {
			validator: prop => prop instanceof Date || prop === null || typeof prop === 'string',
		},
		chooseDateLabel: {
			type: String,
			default() {
				return i18n.global.t('input.datepicker.chooseDate')
			},
		},
		disabled: {
			type: Boolean,
			default: false,
		},
	},
	emits: ['update:modelValue', 'change', 'close', 'close-on-change'],
	mounted() {
		document.addEventListener('click', this.hideDatePopup)
	},
	beforeUnmount() {
		document.removeEventListener('click', this.hideDatePopup)
	},
	watch: {
		modelValue: {
			handler: 'setDateValue',
			immediate: true,
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
		// Since flatpickr dates are strings, we need to convert them to native date objects.
		// To make that work, we need a separate variable since flatpickr does not have a change event.
		flatPickrDate: {
			set(newValue) {
				this.date = createDateFromString(newValue)
				this.updateData()
			},
			get() {
				if (!this.date) {
					return ''
				}

				return format(this.date, 'yyy-LL-dd H:mm')
			},
		},
	},
	methods: {
		setDateValue(newVal) {
			if (newVal === null) {
				this.date = null
				return
			}
			this.date = createDateFromString(newVal)
		},
		updateData() {
			this.changed = true
			this.$emit('update:modelValue', this.date)
			this.$emit('change', this.date)
		},
		toggleDatePopup() {
			if (this.disabled) {
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
			// Kind of dirty, but the timeout allows us to enter a time and click on "confirm" without
			// having to click on another input field before it is actually used.
			setTimeout(() => {
				this.show = false
				this.$emit('close', this.changed)
				if (this.changed) {
					this.changed = false
					this.$emit('close-on-change', this.changed)
				}
			}, 200)
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
})
</script>

<style lang="scss" scoped>
.datepicker {
	input.input {
		display: none;
	}

	&.disabled a {
		cursor: default;
	}

	.datepicker-popup {
		position: absolute;
		z-index: 99;
		width: 320px;
		background: var(--white);
		border-radius: $radius;
		box-shadow: $shadow;

		@media screen and (max-width: ($tablet)) {
			width: calc(100vw - 5rem);
		}

		a:not(.button) {
			display: flex;
			align-items: center;
			padding: 0 .5rem;
			width: 100%;
			height: 2.25rem;
			color: var(--text);
			transition: all $transition;

			&:first-child {
				border-radius: $radius $radius 0 0;
			}

			&:hover {
				background: var(--light);
			}

			.text {
				width: 100%;
				font-size: .85rem;
				display: flex;
				justify-content: space-between;
				padding-right: .25rem;

				.weekday {
					color: var(--text-light);
					text-transform: capitalize;
				}
			}

			.icon {
				width: 2rem;
				text-align: center;
			}
		}

		a.button {
			margin: 1rem;
			width: calc(100% - 2rem);
		}

		:deep(.flatpickr-calendar) {
			margin: 0 auto 8px;
			box-shadow: none;
		}
	}
}
</style>