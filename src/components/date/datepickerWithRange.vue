<template>
	<div class="datepicker-with-range-container">
		<popup>
			<template #trigger="{toggle}">
				<slot name="trigger" :toggle="toggle">
					<x-button @click.prevent.stop="toggle()" type="secondary" :shadow="false" class="mb-2">
						{{ $t('task.show.select') }}
					</x-button>
				</slot>
			</template>
			<template #content="{isOpen}">
				<div class="datepicker-with-range" :class="{'is-open': isOpen}">
					<div class="selections">
						<button
							v-for="(value, text) in dateRanges"
							:key="text"
							@click="setDateRange(value)"
							:class="{'is-active': dateRange === value}">
							{{ $t(text) }}
						</button>
						<button @click="setDateRange('')" :class="{'is-active': customRangeActive}">
							{{ $t('misc.custom') }}
						</button>
					</div>
					<div class="flatpickr-container">
						<flat-pickr
							:config="flatPickerConfig"
							v-model="dateRange"
						/>
					</div>
				</div>
			</template>
		</popup>
	</div>
</template>

<script lang="ts" setup>
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {computed, Ref, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {store} from '@/store'
import {format} from 'date-fns'
import Popup from '@/components/misc/popup'

const {t} = useI18n()

const emit = defineEmits(['dateChanged'])

// FIXME: This seems to always contain the default value - that breaks the picker
const weekStart = computed<number>(() => store.state.auth.settings.weekStart ?? 0)
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: false,
	inline: true,
	mode: 'range',
	locale: {
		firstDayOf7Days: weekStart.value,
	},
}))

const dateRange = ref('')

watch(
	() => dateRange.value,
	(newVal: string | null) => {
		if (newVal === null) {
			return
		}

		const [fromDate, toDate] = newVal.split(' to ')

		if (typeof fromDate === 'undefined' || typeof toDate === 'undefined') {
			return
		}

		emit('dateChanged', {
			dateFrom: fromDate,
			dateTo: toDate,
		})
	},
)

function setDateRange(range: string) {
	dateRange.value = range
}

const dateRanges = {
	// Still using strings for the range instead of an array or object to keep it compatible with the dates from flatpickr
	'task.show.today': 'now/d to now/d+1d',
	'task.show.thisWeek': 'now/w to now/w+1w',
	'task.show.nextWeek': 'now/w+1w to now/w+2w',
	'task.show.next7Days': 'now to now+7d',
	'task.show.thisMonth': 'now/M to now/M+1M',
	'task.show.nextMonth': 'now/M+1M to now/M+2M',
	'task.show.next30Days': 'now to now+30d',
}

const customRangeActive = computed<Boolean>(() => {
	return !Object.values(dateRanges).includes(dateRange.value)
})
</script>

<style lang="scss" scoped>
.datepicker-with-range-container {
	position: relative;

	:deep(.popup) {
		z-index: 10;
		margin-top: 1rem;
		border-radius: $radius;
		border: 1px solid var(--grey-200);
		background-color: var(--white);
		box-shadow: $shadow;

		&.is-open {
			width: 500px;
			height: 320px;
		}
	}
}

.datepicker-with-range {
	display: flex;
	width: 100%;
	height: 100%;
	position: absolute;

	:deep(.flatpickr-calendar) {
		margin: 0 auto 8px;
		box-shadow: none;
	}
}

.flatpickr-container {
	width: 70%;
	border-left: 1px solid var(--grey-200);

	:deep(input.input) {
		display: none;
	}
}

.selections {
	width: 30%;
	display: flex;
	flex-direction: column;
	padding-top: .5rem;

	button {
		display: block;
		width: 100%;
		text-align: left;
		padding: .5rem 1rem;
		transition: $transition;
		font-size: .9rem;
		color: var(--text);
		background: transparent;
		border: 0;
		cursor: pointer;

		&.is-active {
			color: var(--primary);
		}

		&:hover, &.is-active {
			background-color: var(--grey-100);
		}
	}
}
</style>
