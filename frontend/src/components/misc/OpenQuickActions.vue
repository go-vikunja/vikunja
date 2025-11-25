<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'
import {useBaseStore} from '@/stores/base'
import {onBeforeUnmount, onMounted} from 'vue'
import {eventToHotkeyString} from '@github/hotkey'
import {isAppleDevice} from '@/helpers/isAppleDevice'

const baseStore = useBaseStore()

// See https://github.com/github/hotkey/discussions/85#discussioncomment-5214660
function openQuickActionsViaHotkey(event) {
	const hotkeyString = eventToHotkeyString(event)
	if (!hotkeyString) return
	
	// On macOS, use Cmd+K (Meta+K), on other platforms use Ctrl+K (Control+K)
	const expectedHotkey = isAppleDevice() ? 'Meta+k' : 'Control+k'
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
