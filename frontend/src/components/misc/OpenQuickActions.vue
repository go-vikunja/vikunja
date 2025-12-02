<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'
import {useBaseStore} from '@/stores/base'
import {onBeforeUnmount, onMounted} from 'vue'
import {eventToHotkeyString} from '@github/hotkey'
import {useShortcutManager} from '@/composables/useShortcutManager'

const baseStore = useBaseStore()
const shortcutManager = useShortcutManager()

// See https://github.com/github/hotkey/discussions/85#discussioncomment-5214660
function openQuickActionsViaHotkey(event) {
	const hotkeyString = eventToHotkeyString(event)
	if (!hotkeyString) return

	const expectedHotkey = shortcutManager.getHotkeyString('general.quickSearch')
	if (hotkeyString !== expectedHotkey) return

	event.preventDefault()

	openQuickActions()
}

onMounted(() => {
	document.addEventListener('keydown', openQuickActionsViaHotkey)
})

onBeforeUnmount(() => {
	document.removeEventListener('keydown', openQuickActionsViaHotkey)
})

function openQuickActions() {
	baseStore.setQuickActionsActive(true)
}
</script>

<template>
	<BaseButton
		class="trigger-button"
		:title="$t('keyboardShortcuts.quickSearch')"
		@click="openQuickActions"
	>
		<Icon icon="search" />
	</BaseButton>
</template>
