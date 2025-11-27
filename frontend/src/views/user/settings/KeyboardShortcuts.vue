<template>
	<div class="keyboard-shortcuts-settings">
		<header>
			<h2>{{ $t('user.settings.keyboardShortcuts.title') }}</h2>
			<p class="help">
				{{ $t('user.settings.keyboardShortcuts.description') }}
			</p>
			<BaseButton
				variant="secondary"
				@click="resetAll"
			>
				{{ $t('user.settings.keyboardShortcuts.resetAll') }}
			</BaseButton>
		</header>

		<!-- Group by category -->
		<section
			v-for="group in shortcutGroups"
			:key="group.category"
			class="shortcut-group"
		>
			<div class="group-header">
				<h3>{{ $t(group.title) }}</h3>
				<BaseButton
					v-if="hasCustomizableInGroup(group)"
					variant="tertiary"
					size="small"
					@click="resetCategory(group.category)"
				>
					{{ $t('user.settings.keyboardShortcuts.resetCategory') }}
				</BaseButton>
			</div>

			<div class="shortcuts-list">
				<ShortcutEditor
					v-for="shortcut in group.shortcuts"
					:key="shortcut.actionId"
					:shortcut="shortcut"
					@update="updateShortcut"
					@reset="resetShortcut"
				/>
			</div>
		</section>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useShortcutManager } from '@/composables/useShortcutManager'
import { ShortcutCategory, type ShortcutGroup } from '@/components/misc/keyboard-shortcuts/shortcuts'
import ShortcutEditor from '@/components/misc/keyboard-shortcuts/ShortcutEditor.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import { success, error } from '@/message'

const { t } = useI18n()
const shortcutManager = useShortcutManager()

const shortcutGroups = shortcutManager.getAllShortcuts()

function hasCustomizableInGroup(group: ShortcutGroup) {
	return group.shortcuts.some(s => s.customizable)
}

async function updateShortcut(actionId: string, keys: string[]) {
	try {
		const result = await shortcutManager.setCustomShortcut(actionId, keys)
		if (!result.valid) {
			error({
				message: t(result.error || 'keyboardShortcuts.errors.unknown'),
			})
		} else {
			success({
				message: t('user.settings.keyboardShortcuts.shortcutUpdated'),
			})
		}
	} catch (e) {
		error(e)
	}
}

async function resetShortcut(actionId: string) {
	try {
		await shortcutManager.resetShortcut(actionId)
		success({
			message: t('user.settings.keyboardShortcuts.shortcutReset'),
		})
	} catch (e) {
		error(e)
	}
}

async function resetCategory(category: ShortcutCategory) {
	try {
		await shortcutManager.resetCategory(category)
		success({
			message: t('user.settings.keyboardShortcuts.categoryReset'),
		})
	} catch (e) {
		error(e)
	}
}

async function resetAll() {
	if (confirm(t('user.settings.keyboardShortcuts.resetAllConfirm'))) {
		try {
			await shortcutManager.resetAll()
			success({
				message: t('user.settings.keyboardShortcuts.allReset'),
			})
		} catch (e) {
			error(e)
		}
	}
}
</script>

<style scoped>
.keyboard-shortcuts-settings {
	max-width: 800px;
}

header {
	margin-bottom: 2rem;
}

header h2 {
	margin-bottom: 0.5rem;
}

header .help {
	margin-bottom: 1rem;
	color: var(--text-light);
}

.shortcut-group {
	margin-bottom: 2rem;
}

.group-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 1rem;
	padding-bottom: 0.5rem;
	border-bottom: 2px solid var(--grey-200);
}

.group-header h3 {
	margin: 0;
	font-size: 1.25rem;
	font-weight: 600;
}

.shortcuts-list {
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: 4px;
}

.shortcuts-list > :last-child {
	border-bottom: none;
}
</style>
