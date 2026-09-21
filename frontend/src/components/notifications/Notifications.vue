<template>
	<div class="notifications">
		<slot
			name="trigger"
			toggle-open="() => showNotifications = !showNotifications"
			:has-unread-notifications="unreadNotifications > 0"
		>
			<BaseButton
				class="trigger-button"
				:aria-expanded="showNotifications"
				@click.stop="showNotifications = !showNotifications"
			>
				<span class="is-sr-only">{{ $t('notification.title') }}</span>
				<span
					v-if="unreadNotifications > 0"
					class="unread-indicator"
				/>
				<Icon icon="bell" />
			</BaseButton>
		</slot>

		<CustomTransition name="fade">
			<div
				v-if="showNotifications"
				ref="popup"
				class="notifications-list"
			>
				<div class="head">
					<span>{{ $t('notification.title') }}</span>
					<div class="actions">
						<BaseButton
							v-if="notifications.length > 0"
							v-tooltip="$t('notification.clearAll')"
							class="action-link"
							:aria-label="$t('notification.clearAll')"
							@click="clearAll"
						>
							<Icon icon="check-double" />
						</BaseButton>
						<BaseButton
							v-tooltip="$t('notification.subscribeFeed')"
							class="action-link"
							:to="{name: 'user.settings.feeds'}"
							@click="showNotifications = false"
						>
							<span class="is-sr-only">{{ $t('notification.subscribeFeed') }}</span>
							<Icon icon="rss" />
						</BaseButton>
					</div>
				</div>
				<div
					v-for="n in notifications"
					:key="n.id"
					class="single-notification"
					:class="{'is-clickable': notificationRoute(n)}"
					@click="() => to(n)"
				>
					<div
						class="read-indicator"
						:class="{'read': parseDateOrNull(n.read_at) !== null}"
					/>
					<User
						v-if="notificationDoer(n)"
						:user="notificationDoer(n)!"
						:show-username="false"
						:avatar-size="16"
					/>
					<div class="detail">
						<div>
							<span
								v-if="notificationDoer(n)"
								class="has-text-weight-bold mie-1"
							>
								{{ getDisplayName(notificationDoer(n)) }}
							</span>
							{{ notificationText(n, authStore.info) }}
						</div>
						<span
							v-tooltip="formatDateLong(n.created)"
							class="created"
						>
							{{ formatDisplayDate(n.created) }}
						</span>
					</div>
				</div>
				<XButton
					v-if="notifications.length > 0 && unreadNotifications > 0"
					variant="tertiary"
					class="mbs-2 is-fullwidth"
					@click="markAllRead"
				>
					{{ $t('notification.markAllRead') }}
				</XButton>
				<p
					v-if="notifications.length === 0"
					class="nothing"
				>
					{{ $t('notification.none') }}<br>
					<span class="explainer">
						{{ $t('notification.explainer') }}
					</span>
				</p>
			</div>
		</CustomTransition>
	</div>
</template>

<script lang="ts" setup>
import {computed, onMounted, onUnmounted, ref} from 'vue'
import {useRouter, isNavigationFailure, NavigationFailureType} from 'vue-router'
import {useQuery} from '@tanstack/vue-query'
import type {DatabaseNotification} from '@/client/generated'
import {notificationsQuery, useMarkNotificationReadMutation, useMarkAllNotificationsReadMutation, useClearNotificationsMutation} from '@/client/queries/notifications'
import {notificationDoer, notificationRoute, notificationText} from '@/helpers/notification'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import User from '@/components/misc/User.vue'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/helpers/user'
import {useAuthStore} from '@/stores/auth'
import {useWebSocket} from '@/composables/useWebSocket'
import XButton from '@/components/input/Button.vue'

const {authenticated} = useWebSocket()
const authStore = useAuthStore()
const router = useRouter()
const {data} = useQuery(computed(() => ({
	...notificationsQuery(),
	refetchInterval: authenticated.value ? false : 10_000,
})))
const readMutation = useMarkNotificationReadMutation()
const readAllMutation = useMarkAllNotificationsReadMutation()
const clearMutation = useClearNotificationsMutation()
const notifications = computed(() => (data.value ?? []).filter(n => n.name))
const unreadNotifications = computed(() => notifications.value.filter(n => !parseDateOrNull(n.read_at)).length)
const showNotifications = ref(false)
const popup = ref<HTMLElement | null>(null)

onMounted(() => document.addEventListener('click', hidePopup))
onUnmounted(() => document.removeEventListener('click', hidePopup))

function hidePopup(e: MouseEvent) {
	if (showNotifications.value && popup.value !== null) {
		closeWhenClickedOutside(e, popup.value, () => showNotifications.value = false)
	}
}

async function to(n: DatabaseNotification) {
	const route = notificationRoute(n)
	if (!route || !n.id) return
	try {
		await readMutation.mutateAsync(n.id)
		showNotifications.value = false
		const failure = await router.push(route)
		if (isNavigationFailure(failure, NavigationFailureType.duplicated)) router.go(0)
	} catch { /* Mutation reports the error. */ }
}

function markAllRead() { readAllMutation.mutate() }
function clearAll() { clearMutation.mutate() }
</script>

<style lang="scss" scoped>
.notifications {
	display: flex;

	.trigger-button {
		inline-size: 100%;
		position: relative;
	}

	.unread-indicator {
		position: absolute;
		inset-block-start: 1rem;
		inset-inline-end: .5rem;
		inline-size: .75rem;
		block-size: .75rem;

		background: var(--primary);
		border-radius: 100%;
		border: 2px solid var(--white);
	}

	.notifications-list {
		position: absolute;
		inset-inline-end: 1rem;
		inset-block-start: calc(100% + 1rem);
		max-block-size: 400px;
		overflow-y: auto;

		background: var(--white);
		inline-size: 350px;
		max-inline-size: calc(100vw - 2rem);
		padding: .75rem .25rem;
		border-radius: $radius;
		box-shadow: var(--shadow-sm);
		font-size: .85rem;

		@media screen and (max-width: $tablet) {
			max-block-size: calc(100vh - 1rem - #{$navbar-height});
		}

		.head {
			font-family: $vikunja-font;
			font-size: 1rem;
			padding: .5rem;
			display: flex;
			align-items: center;
			justify-content: space-between;

			.actions {
				display: flex;
				align-items: center;
				gap: .5rem;
			}

			.action-link {
				color: var(--grey-500);
				transition: color $transition;

				&:hover,
				&:focus {
					color: var(--primary);
				}
			}
		}

		.single-notification {
			display: flex;
			align-items: center;
			padding: 0.25rem 0;

			transition: background-color $transition;

			&.is-clickable {
				cursor: pointer;
			}

			&:hover {
				background: var(--grey-100);
				border-radius: $radius;
			}

			.read-indicator {
				inline-size: .35rem;
				block-size: .35rem;
				background: var(--primary);
				border-radius: 100%;
				margin: 0 .5rem;
				flex-shrink: 0;

				&.read {
					background: transparent;
				}
			}

			.user {
				display: inline-flex;
				align-items: center;
				inline-size: auto;
				margin: 0 .5rem;

				span {
					font-family: $family-sans-serif;
				}

				.avatar {
					block-size: 16px;
				}

				img {
					margin-inline-end: 0;
				}
			}

			.created {
				color: var(--grey-400);
			}

			&:last-child {
				margin-block-end: .25rem;
			}

			a {
				color: var(--grey-800);
			}
		}

		.nothing {
			text-align: center;
			padding: 1rem 0;
			color: var(--grey-500);

			.explainer {
				font-size: .75rem;
			}
		}
	}
}
</style>
