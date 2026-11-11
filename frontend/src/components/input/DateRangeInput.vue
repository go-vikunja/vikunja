<template>
	<div class="date-range-input">
		<Popup
			v-model:open="isPopupOpen"
			placement="bottom-start"
			:anchor="triggerEl"
			sheet-on-mobile
			:sheet-title="placeholder"
		>
			<template #trigger="{toggle, isOpen}">
				<button
					:id="inputId"
					ref="triggerEl"
					type="button"
					class="input date-range-input__trigger"
					:aria-expanded="isOpen"
					@click.stop="toggle"
				>
					<span v-if="modelValue.start && modelValue.end">
						{{ formatDate(modelValue.start, 'll') }} – {{ formatDate(modelValue.end, 'll') }}
					</span>
					<span
						v-else
						class="date-range-input__placeholder"
					>{{ placeholder }}</span>
				</button>
			</template>
			<template #content>
				<div class="date-range-input__popup">
					<CalendarMonth
						mode="range"
						:range-start="rangeStart"
						:range-end="rangeEnd"
						@pick="pickDay"
					/>
				</div>
			</template>
		</Popup>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, useId} from 'vue'

import Popup from '@/components/misc/Popup.vue'
import CalendarMonth from '@/components/input/datepicker/CalendarMonth.vue'

import {useRangePick} from '@/composables/useRangePick'
import {formatDate} from '@/helpers/time/formatDate'

const props = withDefaults(defineProps<{
	modelValue: {start: Date | null, end: Date | null}
	placeholder?: string
	id?: string
}>(), {
	placeholder: '',
})

const emit = defineEmits<{
	'update:modelValue': [value: {start: Date, end: Date}]
}>()

const fallbackId = useId()
const inputId = computed(() => props.id ?? fallbackId)

const triggerEl = ref<HTMLElement | null>(null)
const isPopupOpen = ref(false)

const {rangeStart, rangeEnd, pickDay} = useRangePick(
	() => props.modelValue.start,
	() => props.modelValue.end,
	(start, end) => emit('update:modelValue', {start, end}),
	isPopupOpen,
)
</script>

<style lang="scss" scoped>
.date-range-input {
	position: relative;
}

.date-range-input__trigger {
	text-align: start;
	cursor: pointer;
}

.date-range-input__placeholder {
	color: var(--input-placeholder-color, var(--grey-400));
}

:deep(.popup) {
	border-radius: $radius;
	border: 1px solid var(--grey-200);
	background-color: var(--white);
	box-shadow: $shadow;

	&.is-open {
		inline-size: 320px;
	}
}

.bottom-sheet .date-range-input__popup {
	padding-block-end: 1.5rem;
}

.date-range-input__popup {
	padding: .75rem 1rem 1rem;
}
</style>
