<template>
	<Modal @close="close()">
		<Card
			class="has-background-white keyboard-shortcuts"
			:shadow="false"
			:title="$t('keyboardShortcuts.title')"
			:show-close="true"
			@close="close()"
		>
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
							:keys="sc.keys"
							:combination="sc.combination && $t(`keyboardShortcuts.${sc.combination}`)"
						/>
					</template>
				</dl>
			</template>
		</Card>
	</Modal>
</template>

<script lang="ts" setup>
import {useBaseStore} from '@/stores/base'

import Shortcut from '@/components/misc/Shortcut.vue'
import Message from '@/components/misc/Message.vue'

import {KEYBOARD_SHORTCUTS as shortcuts} from './shortcuts'

function close() {
	useBaseStore().setKeyboardShortcutsActive(false)
}
</script>

<style scoped>
.keyboard-shortcuts {
	text-align: start;
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
