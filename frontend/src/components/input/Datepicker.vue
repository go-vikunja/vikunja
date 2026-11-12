<template>
	<div class="datepicker">
		<Popup
			ref="popupComponent"
			v-model:open="show"
			placement="bottom-start"
			:anchor="triggerEl"
			sheet-on-mobile
			:sheet-title="dialogTitle"
		>
			<template #trigger="{toggle}">
				<SimpleButton
					:id="triggerId"
					ref="triggerButton"
					v-tooltip="date ? formatDateLong(date) : undefined"
					class="show"
					:aria-labelledby="ariaLabelledby ? `${ariaLabelledby} ${triggerId}` : undefined"
					:disabled="disabled || undefined"
					@click.stop="toggle()"
				>
					<i v-if="date === null && emptyLabel !== ''">{{ emptyLabel }}</i>
					<template v-else>
						{{ date === null ? chooseDateLabel : formatDisplayDate(date) }}
					</template>
				</SimpleButton>
			</template>
			<template #header-action>
				<BaseButton
					class="bottom-sheet__header-action"
					@click.stop="clear"
				>
					{{ $t('input.datepicker.clear') }}
				</BaseButton>
			</template>
			<template #content="{isOpen}">
				<DatepickerPanel
					v-if="isOpen"
					:model-value="date"
					:show-shortcuts="showShortcuts"
					:min-date="minDate"
					:role="isMobile ? undefined : 'dialog'"
					:aria-label="isMobile ? undefined : dialogTitle"
					tabindex="-1"
					@update:modelValue="setDate"
					@confirm="close"
				/>
			</template>
		</Popup>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount, ref, toRef, watch, nextTick, useId} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import Popup from '@/components/misc/Popup.vue'
import DatepickerPanel from '@/components/input/DatepickerPanel.vue'
import SimpleButton from '@/components/input/SimpleButton.vue'

import {formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'
import {createDateFromString} from '@/helpers/time/createDateFromString'
import {useIsMobile} from '@/composables/useIsMobile'

const props = withDefaults(defineProps<{
	modelValue: Date | null | string,
	chooseDateLabel?: string,
	title?: string,
	disabled?: boolean,
	showShortcuts?: boolean,
	// When the value is null, show this (italic) instead of chooseDateLabel.
	emptyLabel?: string,
	minDate?: Date | null,
	ariaLabelledby?: string,
}>(), {
	chooseDateLabel: () => {
		const {t} = useI18n({useScope: 'global'})
		return t('input.datepicker.chooseDate')
	},
	title: '',
	disabled: false,
	showShortcuts: true,
	emptyLabel: '',
	minDate: null,
	ariaLabelledby: undefined,
})

const emit = defineEmits<{
	'update:modelValue': [value: Date | null],
	'close': [value: boolean],
	'closeOnChange': [value: boolean],
}>()

const isMobile = useIsMobile()

const dialogTitle = computed(() => props.title || props.chooseDateLabel)

const date = ref<Date | null>(null)
const show = ref(false)
const changed = ref(false)

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

function setDate(value: Date | null) {
	date.value = value
	updateData()
}

function clear() {
	setDate(null)
}

const triggerId = useId()
const triggerButton = ref<InstanceType<typeof SimpleButton> | null>(null)
const triggerEl = computed<HTMLElement | null>(() => triggerButton.value?.$el ?? null)

const popupComponent = ref<InstanceType<typeof Popup> | null>(null)

async function focusTrigger() {
	await nextTick()
	if (!isMobile.value) {
		triggerButton.value?.focus()
	}
}

watch(show, async (isOpen) => {
	forgetOpenRequest()
	if (!isOpen) {
		emitClose()
		return
	}
	await focusTrigger()
})

// `show` can still be true just after the browser light-dismissed the popup: Popup reports that close only
// once hiding has finished. Setting it again would change nothing, so the popup is asked to open itself.
function showPopup() {
	if (!show.value) {
		show.value = true
		return
	}
	forgetOpenRequest()
	popupComponent.value?.openPopup()
	focusTrigger()
}

function close() {
	forgetOpenRequest()
	show.value = false
}

// A keyboard shortcut can ask while the picker is disabled for a moment (the task is saving). Nothing on
// screen says why it did nothing, so the request is kept until the picker is enabled again, unless the
// person has moved on: any key or pointer press, or ten seconds, drops it.
const OPEN_REQUEST_TTL_MS = 10_000
let openRequested = false
let openRequestTimer: ReturnType<typeof setTimeout> | undefined

function forgetOpenRequest() {
	if (!openRequested) {
		return
	}
	openRequested = false
	clearTimeout(openRequestTimer)
	window.removeEventListener('keydown', forgetOpenRequest, true)
	window.removeEventListener('pointerdown', forgetOpenRequest, true)
}

function rememberOpenRequest() {
	if (openRequested) {
		return
	}
	openRequested = true
	openRequestTimer = setTimeout(forgetOpenRequest, OPEN_REQUEST_TTL_MS)
	window.addEventListener('keydown', forgetOpenRequest, true)
	window.addEventListener('pointerdown', forgetOpenRequest, true)
}

watch(() => props.disabled, disabled => {
	if (!disabled && openRequested) {
		forgetOpenRequest()
		showPopup()
	}
})

onBeforeUnmount(forgetOpenRequest)

function open() {
	if (props.disabled) {
		rememberOpenRequest()
		return
	}
	showPopup()
}

function emitClose() {
	emit('close', changed.value)
	if (changed.value) {
		changed.value = false
		emit('closeOnChange', changed.value)
	}
}

defineExpose({
	open,
})
</script>
