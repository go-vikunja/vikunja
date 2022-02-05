<template>
	<modal @close="close()">
		<card class="has-background-white has-no-shadow keyboard-shortcuts" :title="$t('keyboardShortcuts.title')">
			<template v-for="(s, i) in shortcuts" :key="i">
				<h3>{{ $t(s.title) }}</h3>

				<message class="mb-4" v-if="s.available">
					{{
						s.available($route)
							? $t('keyboardShortcuts.currentPageOnly')
							: $t('keyboardShortcuts.allPages')
					}}
				</message>

				<dl class="shortcut-list">
					<template v-for="(sc, si) in s.shortcuts" :key="si">
						<dt class="shortcut-title">{{ $t(sc.title) }}</dt>
						<shortcut
							class="shortcut-keys"
							is="dd"
							:keys="sc.keys"
							:combination="sc.combination && $t(`keyboardShortcuts.${sc.combination}`)"
						/>
					</template>
				</dl>
			</template>
		</card>
	</modal>
</template>

<script lang="ts" setup>
import {useStore} from 'vuex'

import Shortcut from '@/components/misc/shortcut.vue'
import Message from '@/components/misc/message.vue'

import {KEYBOARD_SHORTCUTS_ACTIVE} from '@/store/mutation-types'
import {KEYBOARD_SHORTCUTS as shortcuts} from './shortcuts'

const store = useStore()
function close() {
	store.commit(KEYBOARD_SHORTCUTS_ACTIVE, false)
}
</script>

<style scoped>
.keyboard-shortcuts {
	text-align: left;
}

.message:not(:last-child) {
	margin-bottom: 1rem;
}

.message-body {
	padding: .75rem;
}

.shortcut-list {
	display: grid;
	grid-template-columns: 2fr 1fr;
}

.shortcut-title {
	margin-bottom: .5rem;
}

.shortcut-keys {
	justify-content: end;
	margin-bottom: .5rem;
}
</style>