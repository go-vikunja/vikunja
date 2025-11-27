<template>
	<div
		class="shortcut-editor"
		:class="{ 'is-disabled': !shortcut.customizable, 'is-editing': isEditing }"
	>
		<div class="shortcut-info">
			<label>{{ $t(shortcut.title) }}</label>
			<span
				v-if="!shortcut.customizable"
				class="tag is-light"
			>
				{{ $t('keyboardShortcuts.fixed') }}
			</span>
		</div>

		<div class="shortcut-input">
			<div
				v-if="!isEditing"
				class="shortcut-display"
			>
				<Shortcut :keys="displayKeys" />
				<BaseButton
					v-if="shortcut.customizable"
					size="small"
					variant="tertiary"
					@click="startEditing"
				>
					{{ $t('misc.edit') }}
				</BaseButton>
			</div>

			<div
				v-else
				class="shortcut-edit"
			>
				<input
					ref="captureInput"
					type="text"
					readonly
					:value="captureDisplay"
					:placeholder="$t('keyboardShortcuts.pressKeys')"
					class="key-capture-input"
					@keydown.prevent="captureKey"
					@blur="cancelEditing"
				>
				<BaseButton
					size="small"
					:disabled="!capturedKeys.length"
					@click="saveShortcut"
				>
					{{ $t('misc.save') }}
				</BaseButton>
				<BaseButton
					size="small"
					variant="tertiary"
					@click="cancelEditing"
				>
					{{ $t('misc.cancel') }}
				</BaseButton>
			</div>

			<BaseButton
				v-if="isCustomized && !isEditing"
				size="small"
				variant="tertiary"
				:title="$t('keyboardShortcuts.resetToDefault')"
				@click="resetToDefault"
			>
				<Icon icon="undo" />
			</BaseButton>
		</div>

		<p
			v-if="validationError"
			class="help is-danger"
		>
			{{ $t(validationError) }}
			<span v-if="conflicts.length">
				{{ conflicts.map(c => $t(c.title)).join(', ') }}
			</span>
		</p>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useShortcutManager } from '@/composables/useShortcutManager'
import { eventToHotkeyString } from '@github/hotkey'
import Shortcut from '@/components/misc/Shortcut.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import { FontAwesomeIcon as Icon } from '@fortawesome/vue-fontawesome'
import type { ShortcutAction } from '@/components/misc/keyboard-shortcuts/shortcuts'

const props = defineProps<{
	shortcut: ShortcutAction
}>()

const emit = defineEmits<{
	update: [actionId: string, keys: string[]]
	reset: [actionId: string]
}>()

const shortcutManager = useShortcutManager()

const isEditing = ref(false)
const capturedKeys = ref<string[]>([])
const validationError = ref<string | null>(null)
const conflicts = ref<ShortcutAction[]>([])
const captureInput = ref<HTMLInputElement>()

const displayKeys = computed(() => {
	return shortcutManager.getShortcut(props.shortcut.actionId) || props.shortcut.keys
})

const isCustomized = computed(() => {
	const current = shortcutManager.getShortcut(props.shortcut.actionId)
	return JSON.stringify(current) !== JSON.stringify(props.shortcut.keys)
})

const captureDisplay = computed(() => {
	return capturedKeys.value.join(' + ')
})

async function startEditing() {
	isEditing.value = true
	capturedKeys.value = []
	validationError.value = null
	conflicts.value = []
	await nextTick()
	captureInput.value?.focus()
}

function captureKey(event: KeyboardEvent) {
	event.preventDefault()

	const hotkeyString = eventToHotkeyString(event)
	if (!hotkeyString) return

	// Parse hotkey string into keys array
	const keys = hotkeyString.includes('+')
		? hotkeyString.split('+')
		: [hotkeyString]

	capturedKeys.value = keys

	// Validate in real-time
	const validation = shortcutManager.validateShortcut(props.shortcut.actionId, keys)
	if (!validation.valid) {
		validationError.value = validation.error || null
		conflicts.value = validation.conflicts || []
	} else {
		validationError.value = null
		conflicts.value = []
	}
}

function saveShortcut() {
	if (!capturedKeys.value.length) return

	const validation = shortcutManager.validateShortcut(props.shortcut.actionId, capturedKeys.value)
	if (!validation.valid) {
		validationError.value = validation.error || null
		conflicts.value = validation.conflicts || []
		return
	}

	emit('update', props.shortcut.actionId, capturedKeys.value)
	isEditing.value = false
	capturedKeys.value = []
}

function cancelEditing() {
	isEditing.value = false
	capturedKeys.value = []
	validationError.value = null
	conflicts.value = []
}

function resetToDefault() {
	emit('reset', props.shortcut.actionId)
}
</script>

<style scoped>
.shortcut-editor {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1rem;
	border-bottom: 1px solid var(--grey-200);
}

.shortcut-editor.is-disabled {
	opacity: 0.6;
	cursor: not-allowed;
}

.shortcut-info {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.shortcut-input {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.shortcut-display {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.shortcut-edit {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.key-capture-input {
	min-width: 200px;
	padding: 0.5rem;
	border: 2px solid var(--primary);
	border-radius: 4px;
	font-family: monospace;
	text-align: center;
}

.help.is-danger {
	color: var(--danger);
	font-size: 0.875rem;
	margin-top: 0.25rem;
}
</style>
