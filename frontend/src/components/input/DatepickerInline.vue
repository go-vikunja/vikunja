<template>
	<BaseButton
		v-if="(new Date()).getHours() < 21"
		class="datepicker__quick-select-date"
		@click.stop="setDate('today')"
	>
		<span class="icon"><Icon :icon="['far', 'calendar-alt']" /></span>
		<span class="text">
			<span>{{ $t('input.datepicker.today') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('today') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('tomorrow')"
	>
		<span class="icon"><Icon :icon="['far', 'sun']" /></span>
		<span class="text">
			<span>{{ $t('input.datepicker.tomorrow') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('tomorrow') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('nextMonday')"
	>
		<span class="icon"><Icon icon="coffee" /></span>
		<span class="text">
			<span>{{ $t('input.datepicker.nextMonday') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('nextMonday') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('thisWeekend')"
	>
		<span class="icon"><Icon icon="cocktail" /></span>
		<span class="text">
			<span>{{ $t('input.datepicker.thisWeekend') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('thisWeekend') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('laterThisWeek')"
	>
		<span class="icon"><Icon icon="chess-knight" /></span>
		<span class="text">
			<span>{{ $t('input.datepicker.laterThisWeek') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('laterThisWeek') }}</span>
		</span>
	</BaseButton>
	<BaseButton
		class="datepicker__quick-select-date"
		@click.stop="setDate('nextWeek')"
	>
		<span class="icon"><Icon icon="forward" /></span>
		<span class="text">
			<span>{{ $t('input.datepicker.nextWeek') }}</span>
			<span class="weekday">{{ getWeekdayFromStringInterval('nextWeek') }}</span>
		</span>
	</BaseButton>

	<div class="flatpickr-container">
		<flat-pickr
			ref="flatPickrRef"
			v-model="flatPickrDate"
			:config="flatPickerConfig"
		/>
	</div>
</template>

<script lang="ts" setup>
import {computed, onBeforeUnmount, onMounted, ref, toRef, watch} from 'vue'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

import BaseButton from '@/components/base/BaseButton.vue'

import {formatDate} from '@/helpers/time/formatDate'
import {calculateDayInterval} from '@/helpers/time/calculateDayInterval'
import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'
import {createDateFromString} from '@/helpers/time/createDateFromString'
import {useI18n} from 'vue-i18n'
import {useFlatpickrLanguage} from '@/helpers/useFlatpickrLanguage'

const props = defineProps<{
	modelValue: Date | null | string
}>()

const emit = defineEmits<{
	'update:modelValue': [Date | null],
}>()

const {t} = useI18n({useScope: 'global'})

const date = ref<Date | null>()
const changed = ref(false)

const modelValue = toRef(props, 'modelValue')
watch(
	modelValue,
	setDateValue,
	{immediate: true},
)

const flatPickrRef = ref<HTMLElement | null>(null)
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: true,
	time_24hr: true,
	inline: true,
	locale: useFlatpickrLanguage().value,
}))

function formatDateToFlatpickrString(date: Date): string {
	const year = date.getFullYear()
	const month = (date.getMonth() + 1).toString().padStart(2, '0')
	const day = date.getDate().toString().padStart(2, '0')
	const hours = date.getHours().toString().padStart(2, '0')
	const minutes = date.getMinutes().toString().padStart(2, '0')
	
	return `${year}-${month}-${day} ${hours}:${minutes}`
}

// Since flatpickr dates are strings, we need to convert them to native date objects.
// To make that work, we need a separate variable since flatpickr does not have a change event.
const flatPickrDate = computed({
	set(newValue: string | Date | null) {
		if (newValue === null) {
			date.value = null
			return
		}

		if (date.value && formatDateToFlatpickrString(date.value) === newValue) {
			return
		}
		date.value = createDateFromString(newValue)
		updateData()
	},
	get() {
		if (!date.value) {
			return ''
		}
		
		return formatDateToFlatpickrString(date.value)
	},
})

onMounted(() => {
	const inputs = flatPickrRef.value?.$el.parentNode.querySelectorAll('.numInputWrapper > input.numInput')
	inputs.forEach(i => {
		i.addEventListener('input', handleFlatpickrInput)
	})
})

onBeforeUnmount(() => {
	const inputs = flatPickrRef.value?.$el.parentNode.querySelectorAll('.numInputWrapper > input.numInput')
	inputs.forEach(i => {
		i.removeEventListener('input', handleFlatpickrInput)
	})
})

// Flatpickr only returns a change event when the value in the input it's referring to changes.
// That means it will usually only trigger when the focus is moved out of the input field.
// This is fine most of the time. However, since we're displaying flatpickr in a popup,
// the whole html dom instance might get destroyed, before the change event had a
// chance to fire. In that case, it would not update the date value. To fix 
// this, we're now listening on every change and bubble them up as soon
// as they happen.
function handleFlatpickrInput(e) {
	const newDate = new Date(date?.value || 'now')
	if (e.target.classList.contains('flatpickr-minute')) {
		newDate.setMinutes(e.target.value)
	}
	if (e.target.classList.contains('flatpickr-hour')) {
		newDate.setHours(e.target.value)
	}
	if (e.target.classList.contains('cur-year')) {
		newDate.setFullYear(e.target.value)
	}
	flatPickrDate.value = newDate
}


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
	const interval = calculateDayInterval(dateString)
	const newDate = new Date()
	newDate.setDate(newDate.getDate() + interval)
	newDate.setHours(calculateNearestHours(newDate))
	newDate.setMinutes(0)
	newDate.setSeconds(0)
	date.value = newDate
	updateData()
}

function getWeekdayFromStringInterval(dateString: string) {
	const interval = calculateDayInterval(dateString)
	const newDate = new Date()
	newDate.setDate(newDate.getDate() + interval)
	return formatDate(newDate, 'ddd')
}
</script>

<style lang="scss" scoped>
.datepicker__quick-select-date {
	display: flex;
	align-items: center;
	padding: 0 .5rem;
	inline-size: 100%;
	block-size: 2.25rem;
	color: var(--text);
	transition: all $transition;

	&:first-child {
		border-radius: $radius $radius 0 0;
	}

	&:hover {
		background: var(--grey-100);
	}

	.text {
		inline-size: 100%;
		font-size: .85rem;
		display: flex;
		justify-content: space-between;
		padding-inline-end: .25rem;

		.weekday {
			color: var(--text-light);
			text-transform: capitalize;
		}
	}

	.icon {
		inline-size: 2rem;
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
