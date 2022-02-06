<template>
	<div class="datepicker-with-range-container">
		<popup>
			<template #trigger="{toggle}">
				<slot name="trigger" :toggle="toggle" :buttonText="buttonText"></slot>
			</template>
			<template #content="{isOpen}">
				<div class="datepicker-with-range" :class="{'is-open': isOpen}">
					<div class="selections">
						<BaseButton @click="setDateRange(null)" :class="{'is-active': customRangeActive}">
							{{ $t('misc.custom') }}
						</BaseButton>
						<BaseButton
							v-for="(value, text) in DATE_RANGES"
							:key="text"
							@click="setDateRange(value)"
							:class="{'is-active': from === value[0] && to === value[1]}">
							{{ $t(`input.datepickerRange.ranges.${text}`) }}
						</BaseButton>
					</div>
					<div class="flatpickr-container input-group">
						<label class="label">
							{{ $t('input.datepickerRange.from') }}
							<div class="field has-addons">
								<div class="control is-fullwidth">
									<input class="input" type="text" v-model="from"/>
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
									<input class="input" type="text" v-model="to"/>
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

						<p>
							{{ $t('input.datepickerRange.math.canuse') }}
							<BaseButton class="has-text-primary" @click="showHowItWorks = true">
								{{ $t('input.datepickerRange.math.learnhow') }}
							</BaseButton>
						</p>

						<modal
							@close="() => showHowItWorks = false"
							:enabled="showHowItWorks"
							transition-name="fade"
							:overflow="true"
							variant="hint-modal"
						>
							<DatemathHelp/>
						</modal>
					</div>
				</div>
			</template>
		</popup>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref, watch} from 'vue'
import {useStore} from 'vuex'
import {useI18n} from 'vue-i18n'

import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

import Popup from '@/components/misc/popup.vue'
import {DATE_RANGES} from '@/components/date/dateRanges'
import BaseButton from '@/components/base/BaseButton.vue'
import DatemathHelp from '@/components/date/datemathHelp.vue'

const store = useStore()
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

const showHowItWorks = ref(false)

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
watch(() => from, inputChanged)
watch(() => to, inputChanged)

function setDateRange(range: string[] | null) {
	if (range === null) {
		from.value = ''
		to.value = ''

		return
	}

	from.value = range[0]
	to.value = range[1]

}

const customRangeActive = computed<Boolean>(() => {
	return !Object.values(DateRanges).some(el => from.value === el[0] && to.value === el[1])
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

.how-it-works-modal {
	font-size: 1rem;

	p {
		display: inline-block !important;
	}
}
</style>
