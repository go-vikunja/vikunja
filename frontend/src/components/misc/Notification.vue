<template>
	<Teleport :to="teleportTarget">
		<Notifications
			position="bottom left"
			:max="2"
			:ignore-duplicates="true"
			class="global-notification"
			role="status"
			aria-live="polite"
		>
			<template #body="{ item, close }">
				<!-- FIXME: overlay whole notification with button and add event listener on that button instead -->
				<div
					class="vue-notification-template vue-notification"
					:class="[
						item.type,
					]"
					@click="close()"
				>
					<div
						v-if="item.title"
						class="notification-title"
					>
						<Icon
							:icon="item.type === 'error' ? 'exclamation-circle' : 'check'"
							class="notification-icon mie-1"
						/>
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

const teleportTarget = ref<string | HTMLElement>('body')
let observer: MutationObserver | null = null

function syncTeleportTarget() {
	const dialogs = document.querySelectorAll<HTMLDialogElement>('dialog.modal-dialog[open]')
	teleportTarget.value = dialogs.item(dialogs.length - 1) ?? 'body'
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

<style lang="scss" scoped>
.global-notification {
	padding: 0 .75rem .75rem;

	// The library clips each toast, which hides its shadow.
	:deep(.vue-notification-wrapper) {
		overflow: visible;
	}
}

// Matches Message.vue instead of the library's solid fills.
.vue-notification {
	z-index: 9999;
	margin: .5rem 0 0;
	border: 1px solid;
	border-radius: $radius;
	color: var(--text);
	box-shadow: var(--shadow-md);

	.notification-title {
		color: var(--text-strong);
	}

	&.success {
		border-color: var(--success);
		background: linear-gradient(hsla(var(--success-h), var(--success-s), var(--success-l), .08), hsla(var(--success-h), var(--success-s), var(--success-l), .08)), var(--white);

		.notification-icon {
			color: var(--success-text);
		}
	}

	&.error {
		border-color: var(--danger);
		background: linear-gradient(hsla(var(--danger-h), var(--danger-s), var(--danger-l), .08), hsla(var(--danger-h), var(--danger-s), var(--danger-l), .08)), var(--white);

		.notification-icon {
			color: var(--danger-text);
		}
	}
}
</style>
