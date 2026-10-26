<template>
	<Teleport :to="teleportTarget">
		<Notifications
			position="bottom left"
			:max="2"
			:ignore-duplicates="true"
			class="global-notification"
			role="status"
			aria-live="polite"
			@start="startActionTimer"
			@destroy="stopActionTimer"
		>
			<template #body="{ item, close }">
				<!-- FIXME: overlay whole notification with button and add event listener on that button instead -->
				<div
					class="vue-notification-template vue-notification"
					:class="[
						item.type,
					]"
					@click="close()"
					@mouseenter="actionTimers.get(item.id)?.hold('hover')"
					@mouseleave="actionTimers.get(item.id)?.release('hover')"
					@focusin="actionTimers.get(item.id)?.hold('focus')"
					@focusout="releaseFocusHold(item, $event)"
				>
					<div
						v-if="item.title"
						class="notification-title"
					>
						{{ item.title }}
					</div>
					<div class="notification-content">
						<template v-if="Array.isArray(item.text)">
							<template
								v-for="(t, k) in item.text"
								:key="k"
							>
								{{ t }}<br>
							</template>
						</template>
						<template v-else>
							{{ item.text }}
						</template>
						<span
							v-if="item.duplicates > 0"
							class="tw:text-xs tw:font-bold tw:ml-1"
						>
							×{{ item.duplicates + 1 }}
						</span>
					</div>
					<div
						v-if="item.data?.actions?.length > 0"
						class="mbs-2 tw:flex tw:justify-end tw:gap-2"
					>
						<XButton
							v-for="(action, i) in item.data.actions"
							:key="'action_' + i"
							:shadow="false"
							class="is-small"
							variant="secondary"
							@click="action.callback"
						>
							{{ action.title }}
						</XButton>
					</div>
				</div>
			</template>
		</Notifications>
	</Teleport>
</template>

<script lang="ts" setup>
import {onBeforeUnmount, onMounted, ref} from 'vue'
import {notify, type NotificationItem} from '@kyvg/vue3-notification'

import {ACTION_TOAST_DURATION, type ToastData} from '@/message'
import {pausableTimeout} from '@/message/pausableTimeout'

const teleportTarget = ref<string | HTMLElement>('body')
let observer: MutationObserver | null = null

function syncTeleportTarget() {
	const dialogs = document.querySelectorAll<HTMLDialogElement>('dialog.modal-dialog[open]')
	teleportTarget.value = dialogs.item(dialogs.length - 1) ?? 'body'
}

const actionTimers = new Map<NotificationItem['id'], ReturnType<typeof pausableTimeout>>()

function hasActions(item: NotificationItem) {
	return ((item.data as Partial<ToastData> | undefined)?.actions?.length ?? 0) > 0
}

function startActionTimer(item: NotificationItem) {
	if (!hasActions(item)) {
		return
	}
	actionTimers.set(item.id, pausableTimeout(ACTION_TOAST_DURATION, () => notify.close(item.id)))
}

function stopActionTimer(item: NotificationItem) {
	actionTimers.get(item.id)?.cancel()
	actionTimers.delete(item.id)
}

function releaseFocusHold(item: NotificationItem, event: FocusEvent) {
	const toast = event.currentTarget as HTMLElement
	if (!toast.contains(event.relatedTarget as Node | null)) {
		actionTimers.get(item.id)?.release('focus')
	}
}

onMounted(() => {
	syncTeleportTarget()
	observer = new MutationObserver(syncTeleportTarget)
	observer.observe(document.body, {
		attributes: true,
		attributeFilter: ['open'],
		childList: true,
		subtree: true,
	})
})

onBeforeUnmount(() => {
	observer?.disconnect()
	observer = null
})
</script>

<style scoped>
.vue-notification {
	z-index: 9999;
}

</style>
