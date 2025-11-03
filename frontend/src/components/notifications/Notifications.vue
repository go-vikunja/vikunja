<template>
	<div class="notifications">
		<slot
			name="trigger"
			toggle-open="() => showNotifications = !showNotifications"
			:has-unread-notifications="unreadNotifications > 0"
		>
			<BaseButton
				class="trigger-button"
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
				<span class="head">{{ $t('notification.title') }}</span>
				<div
					v-for="(n, index) in notifications"
					:key="n.id"
					class="single-notification"
				>
					<div
						class="read-indicator"
						:class="{'read': n.readAt !== null}"
					/>
					<User
						v-if="n.notification.doer"
						:user="n.notification.doer"
						:show-username="false"
						:avatar-size="16"
					/>
					<div class="detail">
						<div>
							<span
								v-if="n.notification.doer"
								class="has-text-weight-bold mie-1"
							>
								{{ getDisplayName(n.notification.doer) }}
							</span>
							<BaseButton
								class="has-text-start"
								@click="() => to(n, index)()"
							>
								{{ n.toText(userInfo) }}
							</BaseButton>
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
import {useRouter} from 'vue-router'

import NotificationService from '@/services/notification'
import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import User from '@/components/misc/User.vue'
import { NOTIFICATION_NAMES as names, type INotification} from '@/modelTypes/INotification'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/models/user'
import {useAuthStore} from '@/stores/auth'
import XButton from '@/components/input/Button.vue'
import {success} from '@/message'
import {useI18n} from 'vue-i18n'

const LOAD_NOTIFICATIONS_INTERVAL = 10000

const authStore = useAuthStore()
const router = useRouter()
const {t} = useI18n()

const allNotifications = ref<INotification[]>([])
const showNotifications = ref(false)
const popup = ref(null)

const unreadNotifications = computed(() => {
	return notifications.value.filter(n => n.readAt === null).length
})
const notifications = computed(() => {
	return allNotifications.value ? allNotifications.value.filter(n => n.name !== '') : []
})
const userInfo = computed(() => authStore.info)

let interval: ReturnType<typeof setInterval>

onMounted(() => {
	loadNotifications()
	document.addEventListener('click', hidePopup)
	document.addEventListener('visibilitychange', loadNotifications)
	interval = setInterval(loadNotifications, LOAD_NOTIFICATIONS_INTERVAL)
})

onUnmounted(() => {
	document.removeEventListener('click', hidePopup)
	document.removeEventListener('visibilitychange', loadNotifications)
	clearInterval(interval)
})

async function loadNotifications() {
	if (document.visibilityState !== 'visible') {
		return
	}
	// We're recreating the notification service here to make sure it uses the latest api user token
	const notificationService = new NotificationService()
	allNotifications.value = await notificationService.getAll()
}

function hidePopup(e) {
	if (showNotifications.value) {
		closeWhenClickedOutside(e, popup.value, () => showNotifications.value = false)
	}
}

function to(n, index) {
	const to = {
		name: '',
		params: {},
	}

	switch (n.name) {
		case names.TASK_COMMENT:
		case names.TASK_ASSIGNED:
		case names.TASK_REMINDER:
		case names.TASK_MENTIONED:
			to.name = 'task.detail'
			to.params.id = n.notification.task.id
			break
		case names.TASK_DELETED:
			// Nothing
			break
		case names.PROJECT_CREATED:
			to.name = 'task.index'
			to.params.projectId = n.notification.project.id
			break
		case names.TEAM_MEMBER_ADDED:
			to.name = 'teams.edit'
			to.params.id = n.notification.team.id
			break
	}

	return async () => {
		if (to.name !== '') {
			router.push(to)
		}

		n.read = true
		const notificationService = new NotificationService()
		allNotifications.value[index] = await notificationService.update(n)
	}
}

async function markAllRead() {
	const notificationService = new NotificationService()
	await notificationService.markAllRead()
	success({message: t('notification.markAllReadSuccess')})
	
	notifications.value.forEach(n => n.readAt = new Date())
}
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
		}

		.single-notification {
			display: flex;
			align-items: center;
			padding: 0.25rem 0;

			transition: background-color $transition;

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
