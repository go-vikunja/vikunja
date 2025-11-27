<template>
	<Modal @close="close()">
		<Card
			class="has-background-white keyboard-shortcuts"
			:shadow="false"
			:title="$t('keyboardShortcuts.title')"
			:show-close="true"
			@close="close()"
		>
			<template #header>
				<div class="help-header">
					<h2>{{ $t('keyboardShortcuts.title') }}</h2>
					<RouterLink
						:to="{ name: 'user.settings.keyboardShortcuts' }"
						class="button is-small"
						@click="close()"
					>
						{{ $t('keyboardShortcuts.customizeShortcuts') }}
					</RouterLink>
				</div>
			</template>
			<template
				v-for="(s, i) in shortcuts"
				:key="i"
			>
				<h3>{{ $t(s.title) }}</h3>

				<Message
					v-if="s.available"
					class="mbe-4"
				>
					{{
						typeof s.available === 'undefined' ?
							$t('keyboardShortcuts.allPages') :
							(
								s.available($route)
									? $t('keyboardShortcuts.currentPageOnly')
									: $t('keyboardShortcuts.somePagesOnly')
							)
					}}
				</Message>

				<dl class="shortcut-list">
					<template
						v-for="(sc, si) in s.shortcuts"
						:key="si"
					>
						<dt class="shortcut-title">
							{{ $t(sc.title) }}
						</dt>
						<Shortcut
							is="dd"
							class="shortcut-keys"
							:keys="getEffectiveKeys(sc)"
							:combination="sc.combination && $t(`keyboardShortcuts.${sc.combination}`)"
						/>
					</template>
				</dl>
			</template>

			<p class="help-text">
				{{ $t('keyboardShortcuts.helpText') }}
			</p>
		</Card>
	</Modal>
</template>

<script lang="ts" setup>
import {useBaseStore} from '@/stores/base'
import {useShortcutManager} from '@/composables/useShortcutManager'

import Shortcut from '@/components/misc/Shortcut.vue'
import Message from '@/components/misc/Message.vue'

import {KEYBOARD_SHORTCUTS as shortcuts} from './shortcuts'
import type {ShortcutAction} from './shortcuts'

const shortcutManager = useShortcutManager()

function close() {
	useBaseStore().setKeyboardShortcutsActive(false)
}

function getEffectiveKeys(shortcut: ShortcutAction): string[] {
	// For shortcuts with actionId, get effective keys from shortcut manager
	if (shortcut.actionId) {
		return shortcutManager.getShortcut(shortcut.actionId) || shortcut.keys
	}
	// Fallback to default keys for backwards compatibility
	return shortcut.keys
}
</script>

<style scoped>
.keyboard-shortcuts {
	text-align: start;
}

.help-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	width: 100%;
}

.help-header h2 {
	margin: 0;
}

.help-text {
	margin-top: 1rem;
	padding-top: 1rem;
	border-top: 1px solid var(--grey-200);
	color: var(--text-light);
	font-size: 0.875rem;
}

.message:not(:last-child) {
	margin-block-end: 1rem;
}

.message-body {
	padding: .75rem;
}

.shortcut-list {
	display: grid;
	grid-template-columns: 2fr 1fr;
}

.shortcut-title {
	margin-block-end: .5rem;
}

.shortcut-keys {
	justify-content: end;
	margin-block-end: .5rem;
}
</style>
