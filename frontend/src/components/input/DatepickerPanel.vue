<template>
	<div
		class="datepicker-popup"
		:class="{'datepicker-popup--no-shortcuts': !showShortcuts}"
	>
		<DatepickerInline
			:model-value="modelValue"
			:show-shortcuts="showShortcuts"
			:shortcuts-layout="isMobile ? 'chips' : 'sidebar'"
			:large="isMobile"
			:min-date="minDate"
			@update:modelValue="value => emit('update:modelValue', value)"
			@confirm="confirm"
		/>

		<div class="datepicker-popup__footer">
			<div class="datepicker-popup__summary">
				<Icon
					:icon="['far', 'calendar-check']"
					class="datepicker-popup__summary-icon"
				/>
				<span>{{ modelValue === null ? $t('input.datepicker.noDate') : formatDate(modelValue, summaryFormat) }}</span>
			</div>
			<!-- On mobile the sheet header carries Clear. -->
			<XButton
				v-if="!isMobile"
				variant="secondary"
				:shadow="false"
				@click.stop="emit('update:modelValue', null)"
			>
				{{ $t('input.datepicker.clear') }}
			</XButton>
			<XButton
				v-cy="'closeDatepicker'"
				class="datepicker__close-button"
				:shadow="false"
				:disabled="confirmDisabled"
				@click.stop="confirm"
			>
				{{ $t('misc.confirm') }}
			</XButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import DatepickerInline from '@/components/input/DatepickerInline.vue'

import {formatDate} from '@/helpers/time/formatDate'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'
import {useIsMobile} from '@/composables/useIsMobile'

const props = withDefaults(defineProps<{
	modelValue: Date | null,
	showShortcuts?: boolean,
	minDate?: Date | null,
	confirmDisabled?: boolean,
}>(), {
	showShortcuts: true,
	minDate: null,
	confirmDisabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: Date | null],
	'confirm': [],
}>()

const isMobile = useIsMobile()

const {store: timeFormat} = useTimeFormat()
const summaryFormat = computed(() => timeFormat.value === TIME_FORMAT.HOURS_24 ? 'ddd, ll HH:mm' : 'ddd, ll hh:mm A')

function confirm() {
	if (!props.confirmDisabled) {
		emit('confirm')
	}
}
</script>

<style lang="scss" scoped>
.datepicker-popup {
	position: relative;
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: $radius-large;
	box-shadow: var(--shadow-md);
	overflow: hidden;
	inline-size: 548px;
	max-inline-size: calc(100vw - 2rem);

	&--no-shortcuts {
		inline-size: 380px;
	}

	.bottom-sheet & {
		inline-size: 100%;
		max-inline-size: none;
		border: 0;
		border-radius: 0;
		box-shadow: none;
	}
}

.datepicker-popup__footer {
	display: flex;
	align-items: center;
	gap: .75rem;
	padding: .75rem 1rem;
	border-block-start: 1px solid var(--grey-200);
	background: var(--grey-50);

	.bottom-sheet & {
		flex-direction: column;
		align-items: stretch;
		gap: .5rem;
		padding-block-end: 1rem;
		background: var(--white);
	}
}

.datepicker-popup__summary {
	display: flex;
	align-items: center;
	gap: .5rem;
	flex: 1;
	min-inline-size: 0;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
	font-size: .85rem;
	font-weight: 600;
	color: var(--grey-700);
	font-variant-numeric: tabular-nums;
}

.datepicker-popup__summary-icon {
	color: var(--primary);
}

.datepicker__close-button {
	.bottom-sheet & {
		min-block-size: 44px;
	}
}
</style>
