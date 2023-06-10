<template>
	<BaseButton
		v-if="(new Date()).getHours() < 21"
		class="datepicker__quick-select-date"
		@click.stop="setDate('today')"
	>
		<span class="icon"><icon :icon="['far', 'calendar-alt']"/></span>
		<span class="text">
			<span>{{ $t('input.datepicker.today') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('today') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('tomorrow')"
	>
		<span class="icon"><icon :icon="['far', 'sun']"/></span>
		<span class="text">
			<span>{{ $t('input.datepicker.tomorrow') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('tomorrow') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('nextMonday')"
	>
		<span class="icon"><icon icon="coffee"/></span>
		<span class="text">
			<span>{{ $t('input.datepicker.nextMonday') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('nextMonday') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('thisWeekend')"
	>
		<span class="icon"><icon icon="cocktail"/></span>
		<span class="text">
			<span>{{ $t('input.datepicker.thisWeekend') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('thisWeekend') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('laterThisWeek')"
	>
		<span class="icon"><icon icon="chess-knight"/></span>
		<span class="text">
			<span>{{ $t('input.datepicker.laterThisWeek') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('laterThisWeek') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('nextWeek')"
	>
		<span class="icon"><icon icon="forward"/></span>
		<span class="text">
			<span>{{ $t('input.datepicker.nextWeek') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('nextWeek') }}</span>
		</span>
	</BaseButton>

	<div class="flatpickr-container">
		<flat-pickr
			:config="flatPickerConfig"
			v-model="flatPickrDate"
		/>
	</div>
</template>

<script lang="ts" setup>
import {ref, toRef, watch, computed, type PropType} from 'vue'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

import BaseButton from '@/components/base/BaseButton.vue'

import {formatDate} from '@/helpers/time/formatDate'
import {calculateDayInterval} from '@/helpers/time/calculateDayInterval'
import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'
import {createDateFromString} from '@/helpers/time/createDateFromString'
import {useAuthStore} from '@/stores/auth'
import {useI18n} from 'vue-i18n'

const props = defineProps({
	modelValue: {
		type: [Date, null, String] as PropType<Date | null | string>,
		validator: prop => prop instanceof Date || prop === null || typeof prop === 'string',
		default: null,
	},
})

const emit = defineEmits(['update:modelValue', 'close-on-change'])

const {t} = useI18n({useScope: 'global'})

const date = ref<Date | null>()
const changed = ref(false)

const modelValue = toRef(props, 'modelValue')
watch(
	modelValue,
	setDateValue,
	{immediate: true},
)

const authStore = useAuthStore()
const weekStart = computed(() => authStore.settings.weekStart)
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: true,
	time_24hr: true,
	inline: true,
	locale: {
		firstDayOfWeek: weekStart.value,
	},
}))

// Since flatpickr dates are strings, we need to convert them to native date objects.
// To make that work, we need a separate variable since flatpickr does not have a change event.
const flatPickrDate = computed({
	set(newValue: string | Date | null) {
		if (newValue === null) {
			date.value = null
			return
		}

		date.value = createDateFromString(newValue)
		updateData()
	},
	get() {
		if (!date.value) {
			return ''
		}

		return formatDate(date.value, 'yyy-LL-dd H:mm')
	},
})


function setDateValue(dateString: string | Date | null) {
	if (dateString === null) {
		date.value = null
		return
	}
	date.value = createDateFromString(dateString)
}

function updateData() {
	changed.value = true
	emit('update:modelValue', date.value)
}

function setDate(dateString: string) {
	if (date.value === null) {
		date.value = new Date()
	}

	const interval = calculateDayInterval(dateString)
	const newDate = new Date()
	newDate.setDate(newDate.getDate() + interval)
	newDate.setHours(calculateNearestHours(newDate))
	newDate.setMinutes(0)
	newDate.setSeconds(0)
	date.value = newDate
	flatPickrDate.value = newDate
	updateData()
}

function getWeekdayFromStringInterval(dateString: string) {
	const interval = calculateDayInterval(dateString)
	const newDate = new Date()
	newDate.setDate(newDate.getDate() + interval)
	return formatDate(newDate, 'E')
}
</script>

<style lang="scss" scoped>
.datepicker__quick-select-date {
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
		background: var(--grey-100);
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

.flatpickr-container {
	:deep(.flatpickr-calendar) {
		margin: 0 auto 8px;
		box-shadow: none;
	}

	:deep(.input) {
		border: none;
	}
}
</style>
