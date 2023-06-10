<template>
	<div class="datepicker">
		<SimpleButton @click.stop="toggleDatePopup" class="show" :disabled="disabled || undefined">
			{{ date === null ? chooseDateLabel : formatDateShort(date) }}
		</SimpleButton>

		<CustomTransition name="fade">
			<div v-if="show" class="datepicker-popup" ref="datepickerPopup">

				<DatepickerInline
					v-model="date"
					@update:model-value="updateData"
				/>

				<x-button
					class="datepicker__close-button"
					:shadow="false"
					@click="close"
					v-cy="'closeDatepicker'"
				>
					{{ $t('misc.confirm') }}
				</x-button>
			</div>
		</CustomTransition>
	</div>
</template>

<script setup lang="ts">
import {ref, onMounted, onBeforeUnmount, toRef, watch, type PropType} from 'vue'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import DatepickerInline from '@/components/input/datepickerInline.vue'
import SimpleButton from '@/components/input/SimpleButton.vue'

import {formatDateShort} from '@/helpers/time/formatDate'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {createDateFromString} from '@/helpers/time/createDateFromString'
import {useI18n} from 'vue-i18n'

const props = defineProps({
	modelValue: {
		type: [Date, null, String] as PropType<Date | null | string>,
		validator: prop => prop instanceof Date || prop === null || typeof prop === 'string',
		default: null,
	},
	chooseDateLabel: {
		type: String,
		default() {
			const {t} = useI18n({useScope: 'global'})
			return t('input.datepicker.chooseDate')
		},
	},
	disabled: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits(['update:modelValue', 'close', 'close-on-change'])

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
			emit('close-on-change', changed.value)
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
	width: 320px;
	background: var(--white);
	border-radius: $radius;
	box-shadow: $shadow;

	@media screen and (max-width: ($tablet)) {
		width: calc(100vw - 5rem);
	}
}

.datepicker__close-button {
	margin: 1rem;
	width: calc(100% - 2rem);
}

:deep(.flatpickr-calendar) {
	margin: 0 auto 8px;
	box-shadow: none;
}
</style>