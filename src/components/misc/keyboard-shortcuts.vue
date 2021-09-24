<template>
	<modal @close="close()">
				<card class="has-background-white has-no-shadow" :title="$t('keyboardShortcuts.title')">
					<div class="message is-primary">
						<div class="message-body">
							{{ $t('keyboardShortcuts.allPages') }}
						</div>
					</div>
					<p>
						<strong>{{ $t('keyboardShortcuts.toggleMenu') }}</strong>
						<shortcut :keys="['ctrl', 'e']"/>
					</p>
					<p>
						<strong>{{ $t('keyboardShortcuts.quickSearch') }}</strong>
						<shortcut :keys="['ctrl', 'k']"/>
					</p>
					<h3>{{ $t('list.kanban.title') }}</h3>
					<div class="message is-primary" v-if="$route.name === 'list.kanban'">
						<div class="message-body">
							{{ $t('keyboardShortcuts.currentPageOnly') }}
						</div>
					</div>
					<p>
						<strong>{{ $t('keyboardShortcuts.task.done') }}</strong>
						<shortcut :keys="['ctrl', 'click']"/>
					</p>
					<h3>{{ $t('keyboardShortcuts.task.title') }}</h3>
					<div
						class="message is-primary"
						v-if="$route.name === 'task.detail' || $route.name === 'task.list.detail' || $route.name === 'task.gantt.detail' || $route.name === 'task.kanban.detail' || $route.name === 'task.detail'">
						<div class="message-body">
							{{ $t('keyboardShortcuts.currentPageOnly') }}
						</div>
					</div>
					<p>
						<strong>{{ $t('keyboardShortcuts.task.assign') }}</strong>
						<shortcut :keys="['a']"/>
					</p>
					<p>
						<strong>{{ $t('keyboardShortcuts.task.labels') }}</strong>
						<shortcut :keys="['l']"/>
					</p>
					<p>
						<strong>{{ $t('keyboardShortcuts.task.dueDate') }}</strong>
						<shortcut :keys="['d']"/>
					</p>
					<p>
						<strong>{{ $t('keyboardShortcuts.task.attachment') }}</strong>
						<shortcut :keys="['f']"/>
					</p>
					<p>
						<strong>{{ $t('keyboardShortcuts.task.related') }}</strong>
						<shortcut :keys="['r']"/>
					</p>
				</card>
	</modal>
</template>

<script>
import {KEYBOARD_SHORTCUTS_ACTIVE} from '@/store/mutation-types'
import Shortcut from '@/components/misc/shortcut.vue'

export default {
	name: 'keyboard-shortcuts',
	components: {Shortcut},
	methods: {
		close() {
			this.$store.commit(KEYBOARD_SHORTCUTS_ACTIVE, false)
		},
	},
}
</script>