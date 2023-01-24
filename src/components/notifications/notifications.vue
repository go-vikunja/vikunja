<template>
	<div class="notifications">
		<slot name="trigger" toggleOpen="() => showNotifications = !showNotifications" :has-unread-notifications="unreadNotifications > 0">
			<BaseButton class="trigger-button" @click.stop="showNotifications = !showNotifications">
				<span class="unread-indicator" v-if="unreadNotifications > 0"></span>
				<icon icon="bell"/>
			</BaseButton>
		</slot>

		<CustomTransition name="fade">
			<div class="notifications-list" v-if="showNotifications" ref="popup">
				<span class="head">{{ $t('notification.title') }}</span>
				<div
					v-for="(n, index) in notifications"
					:key="n.id"
					class="single-notification"
				>
					<div class="read-indicator" :class="{'read': n.readAt !== null}"></div>
					<user
						:user="n.notification.doer"
						:show-username="false"
						:avatar-size="16"
						v-if="n.notification.doer"/>
					<div class="detail">
						<div>
							<span class="has-text-weight-bold mr-1" v-if="n.notification.doer">
								{{ getDisplayName(n.notification.doer) }}
							</span>
							<BaseButton @click="() => to(n, index)()">
								{{ n.toText(userInfo) }}
							</BaseButton>
						</div>
						<span class="created" v-tooltip="formatDateLong(n.created)">
							{{ formatDateSince(n.created) }}
						</span>
					</div>
				</div>
				<p class="nothing" v-if="notifications.length === 0">
					{{ $t('notification.none') }}<br/>
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
import User from '@/components/misc/user.vue'
import { NOTIFICATION_NAMES as names, type INotification} from '@/modelTypes/INotification'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {formatDateLong, formatDateSince} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/models/user'
import {useAuthStore} from '@/stores/auth'

const LOAD_NOTIFICATIONS_INTERVAL = 10000

const authStore = useAuthStore()
const router = useRouter()

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
	interval = setInterval(loadNotifications, LOAD_NOTIFICATIONS_INTERVAL)
})

onUnmounted(() => {
	document.removeEventListener('click', hidePopup)
	clearInterval(interval)
})

async function loadNotifications() {
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
			to.name = 'task.detail'
			to.params.id = n.notification.task.id
			break
		case names.TASK_DELETED:
			// Nothing
			break
		case names.LIST_CREATED:
			to.name = 'task.index'
			to.params.listId = n.notification.list.id
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
</script>

<style lang="scss" scoped>
.notifications {
	display: flex;

	.trigger-button {
		width: 100%;
	}

	.unread-indicator {
		position: absolute;
		top: .75rem;
		right: 1.15rem;
		width: .75rem;
		height: .75rem;

		background: var(--primary);
		border-radius: 100%;
		border: 2px solid var(--white);
	}

	.notifications-list {
		position: absolute;
		right: 1rem;
		top: calc(100% + 1rem);
		max-height: 400px;
		overflow-y: auto;

		background: var(--white);
		width: 350px;
		max-width: calc(100vw - 2rem);
		padding: .75rem .25rem;
		border-radius: $radius;
		box-shadow: var(--shadow-sm);
		font-size: .85rem;

		@media screen and (max-width: $tablet) {
			max-height: calc(100vh - 1rem - #{$navbar-height});
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
				width: .35rem;
				height: .35rem;
				background: var(--primary);
				border-radius: 100%;
				margin-left: .5rem;

				&.read {
					background: transparent;
				}
			}

			.user {
				display: inline-flex;
				align-items: center;
				width: auto;
				margin: 0 .5rem;

				span {
					font-family: $family-sans-serif;
				}

				.avatar {
					height: 16px;
				}

				img {
					margin-right: 0;
				}
			}

			.created {
				color: var(--grey-400);
			}

			&:last-child {
				margin-bottom: .25rem;
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