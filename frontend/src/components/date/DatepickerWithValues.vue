<template>
	<div class="datepicker-with-range-container">
		<Popup
			:open="open"
			:ignore-click-classes="ignoreClickClasses"
			@update:open="(open) => !open && $emit('update:open', false)"
		>
			<template #content="{isOpen}">
				<div
					class="datepicker-with-range"
					:class="{'is-open': isOpen}"
				>
					<div class="selections">
						<BaseButton
							:class="{'is-active': customRangeActive}"
							@click="setDate(null)"
						>
							{{ $t('misc.custom') }}
						</BaseButton>
						<BaseButton
							v-for="(value, text) in DATE_VALUES"
							:key="text"
							:class="{'is-active': date === value}"
							@click="setDate(value)"
						>
							{{ $t(`input.datepickerRange.values.${text}`) }}
						</BaseButton>
					</div>
					<div class="flatpickr-container input-group">
						<label class="label">
							{{ $t('input.datepickerRange.date') }}
							<div class="field has-addons">
								<div class="control is-fullwidth">
									<input
										v-model="date"
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
							v-model="flatpickrDate"
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
import {DATE_VALUES} from '@/components/date/dateRanges'
import BaseButton from '@/components/base/BaseButton.vue'
import DatemathHelp from '@/components/date/DatemathHelp.vue'
import {useFlatpickrLanguage} from '@/helpers/useFlatpickrLanguage'

const props = withDefaults(defineProps<{
	modelValue: string | Date | null,
	open?: boolean
	ignoreClickClasses?: string[]
}>(), {
	open: false,
	ignoreClickClasses: () => [],
})

const emit = defineEmits<{
	'update:modelValue': [value: string | Date | null],
	'update:open': [open: boolean],
}>()

const {t} = useI18n({useScope: 'global'})

const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: false,
	wrap: true,
	locale: useFlatpickrLanguage().value,
}))

const showHowItWorks = ref(false)

const flatpickrDate = ref('')

const date = ref<string|Date>('')

watch(
	() => props.modelValue,
	newValue => {
		date.value = newValue
		// Only set the date back to flatpickr when it's an actual date.
		// Otherwise flatpickr runs in an endless loop and slows down the browser.
		const parsed = parseDateOrString(date.value, false)
		if (parsed instanceof Date) {
			flatpickrDate.value = date.value
		}
	},
)

function emitChanged() {
	emit('update:modelValue', date.value === '' ? null : date.value)
}

watch(
	() => flatpickrDate.value,
	(newVal: string | null) => {
		if (newVal === null) {
			return
		}

		date.value = newVal

		emitChanged()
	},
)
watch(() => date.value, emitChanged)

function setDate(range: string | null) {
	if (range === null) {
		date.value = ''

		return
	}

	date.value = range
}

const customRangeActive = computed<boolean>(() => {
	return !Object.values(DATE_VALUES).some(d => date.value === d)
})
</script>

<style lang="scss" scoped>
.datepicker-with-range-container {
	position: relative;
}

:deep(.popup) {
	z-index: 10;
	margin-block-start: 1rem;
	border-radius: $radius;
	border: 1px solid var(--grey-200);
	background-color: var(--white);
	box-shadow: $shadow;

	&.is-open {
		inline-size: 500px;
		block-size: 320px;
	}
}

.datepicker-with-range {
	display: flex;
	inline-size: 100%;
	block-size: 100%;
	position: absolute;
}

:deep(.flatpickr-calendar) {
	margin: 0 auto 8px;
	box-shadow: none;
}

.flatpickr-container {
	inline-size: 70%;
	border-inline-start: 1px solid var(--grey-200);
	padding: 1rem;
	font-size: .9rem;

	// Flatpickr has no option to use it without an input field so we're hiding it instead
	:deep(input.form-control.input) {
		block-size: 0;
		padding: 0;
		border: 0;
	}

	.field .control :deep(.button) {
		border: 1px solid var(--input-border-color);
		block-size: 2.25rem;

		&:hover {
			border: 1px solid var(--input-hover-border-color);
		}
	}

	.label, .input, :deep(.button) {
		font-size: .9rem;
	}
}

.selections {
	inline-size: 30%;
	display: flex;
	flex-direction: column;
	padding-block-start: .5rem;
	overflow-y: scroll;

	button {
		display: block;
		inline-size: 100%;
		text-align: start;
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
