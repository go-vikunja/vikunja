<template>
	<modal @close="close()">
		<card class="has-background-white has-no-shadow" :title="$t('keyboardShortcuts.title')">
			<template v-for="(s, i) in shortcuts" :key="i">
				<h3>{{ $t(s.title) }}</h3>

				<div class="message is-primary">
					<div class="message-body">
						{{
							s.available($route) ? $t('keyboardShortcuts.currentPageOnly') : $t('keyboardShortcuts.allPages')
						}}
					</div>
				</div>

				<dl>
					<template v-for="(sc, si) in s.shortcuts" :key="si">
						<dt>{{ $t(sc.title) }}</dt>
						<shortcut
							is="dd"
							:keys="sc.keys"
							:combination="typeof sc.combination !== 'undefined' ? $t(`keyboardShortcuts.${sc.combination}`) : null"/>
					</template>
				</dl>
			</template>
		</card>
	</modal>
</template>

<script>
import {KEYBOARD_SHORTCUTS_ACTIVE} from '@/store/mutation-types'
import Shortcut from '@/components/misc/shortcut.vue'
import {KEYBOARD_SHORTCUTS} from './shortcuts'

export default {
	name: 'keyboard-shortcuts',
	components: {Shortcut},
	data() {
		return {
			shortcuts: KEYBOARD_SHORTCUTS,
		}
	},
	methods: {
		close() {
			this.$store.commit(KEYBOARD_SHORTCUTS_ACTIVE, false)
		},
	},
}
</script>

<style>
dt {
	font-weight: bold;
}
</style>