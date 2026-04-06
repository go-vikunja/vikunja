<template>
	<div class="shortcut-recorder">
		<button
			class="input recorder-button"
			:class="{'is-recording': recording}"
			@click="startRecording"
			@keydown.prevent="onKeyDown"
			@blur="stopRecording"
		>
			<template v-if="recording">
				<span class="recording-hint">{{ $t('user.settings.desktop.shortcutRecorderRecording') }}</span>
			</template>
			<template v-else-if="displayKeys.length > 0">
				<kbd
					v-for="(key, i) in displayKeys"
					:key="i"
				>
					{{ key }}
				</kbd>
			</template>
			<template v-else>
				<span class="placeholder">{{ $t('user.settings.desktop.shortcutRecorderPlaceholder') }}</span>
			</template>
		</button>
		<BaseButton
			v-if="modelValue"
			class="clear-button"
			@click="clear"
		>
			<Icon icon="times" />
		</BaseButton>
	</div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'

const props = defineProps<{
	modelValue: string,
}>()

const emit = defineEmits<{
	'update:modelValue': [value: string],
}>()

const recording = ref(false)

const isMac = navigator.platform.toUpperCase().includes('MAC')

// Map KeyboardEvent properties to Electron accelerator format
function eventToAccelerator(event: KeyboardEvent): string | null {
	if (['Control', 'Alt', 'Shift', 'Meta'].includes(event.key)) {
		return null
	}

	const parts: string[] = []

	if (event.ctrlKey || event.metaKey) parts.push('CmdOrCtrl')
	if (event.altKey) parts.push('Alt')
	if (event.shiftKey) parts.push('Shift')

	// Need at least one modifier for a global shortcut
	if (parts.length === 0) return null

	const key = mapKey(event)
	if (key) parts.push(key)
	else return null

	return parts.join('+')
}

function mapKey(event: KeyboardEvent): string | null {
	// Letters
	if (/^Key[A-Z]$/.test(event.code)) {
		return event.code.slice(3)
	}
	// Digits
	if (/^Digit[0-9]$/.test(event.code)) {
		return event.code.slice(5)
	}
	// Function keys
	if (/^F\d{1,2}$/.test(event.code)) {
		return event.code
	}
	// Special keys
	const specialMap: Record<string, string> = {
		Space: 'Space',
		Enter: 'Enter',
		Backspace: 'Backspace',
		Delete: 'Delete',
		Tab: 'Tab',
		Escape: 'Escape',
		ArrowUp: 'Up',
		ArrowDown: 'Down',
		ArrowLeft: 'Left',
		ArrowRight: 'Right',
		Home: 'Home',
		End: 'End',
		PageUp: 'PageUp',
		PageDown: 'PageDown',
		Minus: '-',
		Equal: '=',
		BracketLeft: '[',
		BracketRight: ']',
		Semicolon: ';',
		Quote: '\'',
		Backquote: '`',
		Backslash: '\\',
		Comma: ',',
		Period: '.',
		Slash: '/',
	}
	return specialMap[event.code] ?? null
}

// Convert Electron accelerator string to display-friendly key names
function acceleratorToDisplayKeys(accelerator: string): string[] {
	if (!accelerator) return []
	return accelerator.split('+').map(part => {
		if (part === 'CmdOrCtrl') return isMac ? '\u2318' : 'Ctrl'
		if (part === 'Shift') return isMac ? '\u21E7' : 'Shift'
		if (part === 'Alt') return isMac ? '\u2325' : 'Alt'
		if (part === 'Space') return '\u2423'
		return part
	})
}

const displayKeys = computed(() => acceleratorToDisplayKeys(props.modelValue))

function startRecording() {
	recording.value = true
}

function stopRecording() {
	recording.value = false
}

function onKeyDown(event: KeyboardEvent) {
	if (!recording.value) {
		startRecording()
	}

	const accelerator = eventToAccelerator(event)
	if (accelerator) {
		emit('update:modelValue', accelerator)
		recording.value = false
	}
}

function clear() {
	emit('update:modelValue', '')
}
</script>

<style lang="scss" scoped>
.shortcut-recorder {
	display: flex;
	align-items: center;
	gap: .5rem;
}

.recorder-button {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	cursor: pointer;
	min-inline-size: 150px;
	text-align: start;

	&.is-recording {
		border-color: var(--primary);
		box-shadow: 0 0 0 0.125em rgba(var(--primary-rgb), 0.25);
	}
}

kbd {
	padding: .1rem .4rem;
	border: 1px solid var(--grey-300);
	background: var(--grey-100);
	border-radius: 3px;
	font-size: .85rem;
	font-family: inherit;
	line-height: 1.5;

	& + kbd {
		margin-inline-start: .15rem;
	}
}

.recording-hint {
	color: var(--primary);
	font-size: .85rem;
}

.placeholder {
	color: var(--grey-400);
}

.clear-button {
	color: var(--grey-500);
	padding: .25rem;

	&:hover {
		color: var(--danger);
	}
}
</style>
