<template>
	<div class="datepicker">
		<SimpleButton
			ref="triggerButton"
			class="show"
			:disabled="disabled || undefined"
			@click.stop="toggleDatePopup"
		>
			<i v-if="date === null && emptyLabel !== ''">{{ emptyLabel }}</i>
			<template v-else>
				{{ date === null ? chooseDateLabel : formatDisplayDate(date) }}
			</template>
		</SimpleButton>

		<CustomTransition name="fade">
			<div
				v-if="show"
				ref="datepickerPopup"
				class="datepicker-popup"
				role="dialog"
				:aria-label="chooseDateLabel"
				tabindex="-1"
				@keydown.esc.stop="closeViaEsc"
			>
				<DatepickerInline
					v-model="date"
					:show-shortcuts="showShortcuts"
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
import {ref, onMounted, onBeforeUnmount, toRef, watch, nextTick} from 'vue'

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
	showShortcuts?: boolean,
	// When the value is null, show this (italic) instead of chooseDateLabel.
	emptyLabel?: string,
}>(), {
	chooseDateLabel: () => {
		const {t} = useI18n({useScope: 'global'})
		return t('input.datepicker.chooseDate')
	},
	disabled: false,
	showShortcuts: true,
	emptyLabel: '',
})

const emit = defineEmits<{
	'update:modelValue': [value: Date | null],
	'close': [value: boolean],
	'closeOnChange': [value: boolean],
}>()

const date = ref<Date | null>(null)
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
	emit('update:modelValue', date.value ?? null)
}

function toggleDatePopup() {
	if (props.disabled) {
		return
	}

	show.value = !show.value
}

const datepickerPopup = ref<HTMLElement | null>(null)
const triggerButton = ref<{$el?: HTMLElement} | null>(null)

// Focus the popup when it opens so the dialog receives focus (ARIA dialog pattern) and its Esc
// handler is reachable. nextTick is required because the popup lives inside a v-if transition.
watch(show, async (isOpen) => {
	if (!isOpen) {
		return
	}
	await nextTick()
	datepickerPopup.value?.focus()
})

function hideDatePopup(e: MouseEvent) {
	if (show.value && datepickerPopup.value) {
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

function closeViaEsc() {
	// Restore focus to the trigger immediately: close() defers show=false by 200ms, but on Esc the
	// popup is about to unmount and focus would otherwise drop to <body>.
	triggerButton.value?.$el?.focus()
	close()
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
