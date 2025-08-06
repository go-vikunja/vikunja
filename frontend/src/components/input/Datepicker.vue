<template>
	<div class="datepicker">
		<SimpleButton
			class="show"
			:disabled="disabled || undefined"
			@click.stop="toggleDatePopup"
		>
			{{ date === null ? chooseDateLabel : formatDisplayDate(date) }}
		</SimpleButton>

		<CustomTransition name="fade">
			<div
				v-if="show"
				ref="datepickerPopup"
				class="datepicker-popup"
			>
				<DatepickerInline
					v-model="date"
					@update:modelValue="updateData"
				/>

				<XButton
					v-cy="'closeDatepicker'"
					class="datepicker__close-button"
					:shadow="false"
					@click="close"
				>
					{{ $t('misc.confirm') }}
				</XButton>
			</div>
		</CustomTransition>
	</div>
</template>

<script setup lang="ts">
import {ref, onMounted, onBeforeUnmount, toRef, watch} from 'vue'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import DatepickerInline from '@/components/input/DatepickerInline.vue'
import SimpleButton from '@/components/input/SimpleButton.vue'

import {formatDisplayDate} from '@/helpers/time/formatDate'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {createDateFromString} from '@/helpers/time/createDateFromString'
import {useI18n} from 'vue-i18n'

const props = withDefaults(defineProps<{
	modelValue: Date | null | string,
	chooseDateLabel?: string,
	disabled?: boolean,
}>(), {
	chooseDateLabel: () => {
		const {t} = useI18n({useScope: 'global'})
		return t('input.datepicker.chooseDate')
	},
	disabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: Date | null],
	'close': [value: boolean],
	'closeOnChange': [value: boolean],
}>()

const date = ref<Date | null>()
const show = ref(false)
const changed = ref(false)

onMounted(() => document.addEventListener('click', hideDatePopup))
onBeforeUnmount(() =>document.removeEventListener('click', hideDatePopup))

const modelValue = toRef(props, 'modelValue')
watch(
	modelValue,
	setDateValue,
	{immediate: true},
)

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

function toggleDatePopup() {
	if (props.disabled) {
		return
	}

	show.value = !show.value
}

const datepickerPopup = ref<HTMLElement | null>(null)
function hideDatePopup(e: MouseEvent) {
	if (show.value) {
		closeWhenClickedOutside(e, datepickerPopup.value, close)
	}
}

function close() {
	// Kind of dirty, but the timeout allows us to enter a time and click on "confirm" without
	// having to click on another input field before it is actually used.
	setTimeout(() => {
		show.value = false
		emit('close', changed.value)
		if (changed.value) {
			changed.value = false
			emit('closeOnChange', changed.value)
		}
	}, 200)
}
</script>

<style lang="scss" scoped>
.datepicker {
	input.input {
		display: none;
	}
}

.datepicker-popup {
	position: absolute;
	z-index: 99;
	inline-size: 320px;
	background: var(--white);
	border-radius: $radius;
	box-shadow: $shadow;

	@media screen and (max-width: ($tablet)) {
		inline-size: calc(100vw - 5rem);
	}
}

.datepicker__close-button {
	margin: 1rem;
	inline-size: calc(100% - 2rem);
}

:deep(.flatpickr-calendar) {
	margin: 0 auto 8px;
	box-shadow: none;
}
</style>
