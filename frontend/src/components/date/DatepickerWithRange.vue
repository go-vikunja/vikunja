<template>
	<div class="datepicker-with-range-container">
		<Popup>
			<template #trigger="{toggle}">
				<slot
					name="trigger"
					:toggle="toggle"
					:button-text="buttonText"
				/>
			</template>
			<template #content="{isOpen}">
				<div
					class="datepicker-with-range"
					:class="{'is-open': isOpen}"
				>
					<div class="selections">
						<BaseButton
							:class="{'is-active': customRangeActive}"
							@click="setDateRange(null)"
						>
							{{ $t('misc.custom') }}
						</BaseButton>
						<BaseButton
							v-for="(value, text) in DATE_RANGES"
							:key="text"
							:class="{'is-active': from === value[0] && to === value[1]}"
							@click="setDateRange(value)"
						>
							{{ $t(`input.datepickerRange.ranges.${text}`) }}
						</BaseButton>
					</div>
					<div class="flatpickr-container input-group">
						<label class="label">
							{{ $t('input.datepickerRange.from') }}
							<div class="field has-addons">
								<div class="control is-fullwidth">
									<input
										v-model="from"
										class="input"
										type="text"
									>
								</div>
								<div class="control">
									<XButton
										icon="calendar"
										variant="secondary"
										data-toggle
									/>
								</div>
							</div>
						</label>
						<label class="label">
							{{ $t('input.datepickerRange.to') }}
							<div class="field has-addons">
								<div class="control is-fullwidth">
									<input
										v-model="to"
										class="input"
										type="text"
									>
								</div>
								<div class="control">
									<XButton
										icon="calendar"
										variant="secondary"
										data-toggle
									/>
								</div>
							</div>
						</label>
						<flat-pickr
							v-model="flatpickrRange"
							:config="flatPickerConfig"
						/>

						<p>
							{{ $t('input.datemathHelp.canuse') }}
						</p>

						<BaseButton
							class="has-text-primary"
							@click="showHowItWorks = true"
						>
							{{ $t('input.datemathHelp.learnhow') }}
						</BaseButton>

						<Modal
							:enabled="showHowItWorks"
							transition-name="fade"
							:overflow="true"
							variant="hint-modal"
							@close="() => showHowItWorks = false"
						>
							<DatemathHelp />
						</Modal>
					</div>
				</div>
			</template>
		</Popup>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {parseDateOrString} from '@/helpers/time/parseDateOrString'

import Popup from '@/components/misc/Popup.vue'
import {DATE_RANGES} from '@/components/date/dateRanges'
import BaseButton from '@/components/base/BaseButton.vue'
import DatemathHelp from '@/components/date/DatemathHelp.vue'
import {useFlatpickrLanguage} from '@/helpers/useFlatpickrLanguage'

const props = defineProps<{
	modelValue: {
		dateFrom: Date | string,
		dateTo: Date | string,
	},
}>()

const emit = defineEmits<{
	'update:modelValue': [value: {
		dateFrom: Date | string,
		dateTo: Date | string
	}]
}>()

const {t} = useI18n({useScope: 'global'})

const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: false,
	wrap: true,
	mode: 'range',
	locale: useFlatpickrLanguage().value,
}))

const showHowItWorks = ref(false)

const flatpickrRange = ref('')

const from = ref('')
const to = ref('')

watch(
	() => props.modelValue,
	newValue => {
		from.value = newValue.dateFrom
		to.value = newValue.dateTo
		// Only set the date back to flatpickr when it's an actual date.
		// Otherwise flatpickr runs in an endless loop and slows down the browser.
		const dateFrom = parseDateOrString(from.value, false)
		const dateTo = parseDateOrString(to.value, false)
		if (dateFrom instanceof Date && dateTo instanceof Date) {
			flatpickrRange.value = `${from.value} to ${to.value}`
		}
	},
)

function emitChanged() {
	const args = {
		dateFrom: from.value === '' ? null : from.value,
		dateTo: to.value === '' ? null : to.value,
	}
	emit('update:modelValue', args)
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
watch(() => from.value, emitChanged)
watch(() => to.value, emitChanged)

function setDateRange(range: string[] | null) {
	if (range === null) {
		from.value = ''
		to.value = ''

		return
	}

	from.value = range[0]
	to.value = range[1]
}

const customRangeActive = computed<boolean>(() => {
	return !Object.values(DATE_RANGES).some(range => from.value === range[0] && to.value === range[1])
})

const buttonText = computed<string>(() => {
	if (from.value !== '' && to.value !== '') {
		return t('input.datepickerRange.fromto', {
			from: from.value,
			to: to.value,
		})
	}

	return t('task.show.select')
})
</script>

<style lang="scss" scoped>
.datepicker-with-range-container {
	position: relative;
}

::v-deep(.popup) {
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

.datepicker-with-range {
	display: flex;
	width: 100%;
	height: 100%;
	position: absolute;
}

::v-deep(.flatpickr-calendar) {
	margin: 0 auto 8px;
	box-shadow: none;
}

.flatpickr-container {
	width: 70%;
	border-left: 1px solid var(--grey-200);
	padding: 1rem;
	font-size: .9rem;

	// Flatpickr has no option to use it without an input field so we're hiding it instead
	::v-deep(input.form-control.input) {
		height: 0;
		padding: 0;
		border: 0;
	}

	.field .control ::v-deep(.button) {
		border: 1px solid var(--input-border-color);
		height: 2.25rem;

		&:hover {
			border: 1px solid var(--input-hover-border-color);
		}
	}

	.label, .input, ::v-deep(.button) {
		font-size: .9rem;
	}
}

.selections {
	width: 30%;
	display: flex;
	flex-direction: column;
	padding-top: .5rem;
	overflow-y: scroll;

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
