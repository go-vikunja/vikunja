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
							:class="{'is-active': from === value[0] && to === value[1]}">
							{{ $t(text) }}
						</button>
						<button @click="setDateRange(null)" :class="{'is-active': customRangeActive}">
							{{ $t('misc.custom') }}
						</button>
					</div>
					<div class="flatpickr-container input-group">
						<label class="label">
							{{ $t('input.datepickerRange.from') }}
							<div class="field has-addons">
								<div class="control is-fullwidth">
									<input class="input" type="text" v-model="from" @change="inputChanged"/>
								</div>
								<div class="control">
									<x-button icon="calendar" variant="secondary" data-toggle/>
								</div>
							</div>
						</label>
						<label class="label">
							{{ $t('input.datepickerRange.to') }}
							<div class="field has-addons">
								<div class="control is-fullwidth">
									<input class="input" type="text" v-model="to" @change="inputChanged"/>
								</div>
								<div class="control">
									<x-button icon="calendar" variant="secondary" data-toggle/>
								</div>
							</div>
						</label>
						<flat-pickr
							:config="flatPickerConfig"
							v-model="flatpickrRange"
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
	wrap: true,
	mode: 'range',
	locale: {
		firstDayOf7Days: weekStart.value,
	},
}))

const flatpickrRange = ref('')

const from = ref('')
const to = ref('')

function emitChanged() {
	emit('dateChanged', {
		dateFrom: from.value === '' ? null : from.value,
		dateTo: to.value === '' ? null : to.value,
	})
}

function inputChanged() {
	flatpickrRange.value = ''
	emitChanged()
}

watch(
	() => flatpickrRange.value,
	(newVal: string | null) => {
		if (newVal === null) {
			return
		}

		const [fromDate, toDate] = newVal.split(' to ')

		if (typeof fromDate === 'undefined' || typeof toDate === 'undefined') {
			return
		}

		from.value = fromDate
		to.value = toDate

		emitChanged()
	},
)

function setDateRange(range: string[] | null) {
	if (range === null) {
		from.value = ''
		to.value = ''
		inputChanged()

		return
	}

	from.value = range[0]
	to.value = range[1]

	inputChanged()
}

const dateRanges = {
	'task.show.today': ['now/d', 'now/d+1d'],
	'task.show.thisWeek': ['now/w', 'now/w+1w'],
	'task.show.nextWeek': ['now/w+1w', 'now/w+2w'],
	'task.show.next7Days': ['now', 'now+7d'],
	'task.show.thisMonth': ['now/M', 'now/M+1M'],
	'task.show.nextMonth': ['now/M+1M', 'now/M+2M'],
	'task.show.next30Days': ['now', 'now+30d'],
}

const customRangeActive = computed<Boolean>(() => {
	return !Object.values(dateRanges).some(el => from.value === el[0] && to.value === el[1])
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
	padding: 1rem;
	font-size: .9rem;

	// Flatpickr has no option to use it without an input field so we're hiding it instead
	:deep(input.form-control.input) {
		height: 0;
		padding: 0;
		border: 0;
	}

	.field .control :deep(.button) {
		border: 1px solid var(--input-border-color);
		height: 2.25rem;

		&:hover {
			border: 1px solid var(--input-hover-border-color);
		}
	}

	.label, .input, :deep(.button) {
		font-size: .9rem;
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
