<template>
	<div class="notifications">
		<div class="is-flex is-justify-content-center">
			<a @click.stop="showNotifications = !showNotifications" class="trigger-button">
				<span class="unread-indicator" v-if="unreadNotifications > 0"></span>
				<icon icon="bell"/>
			</a>
		</div>

		<transition name="fade">
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
								{{ n.notification.doer.getDisplayName() }}
							</span>
							<a @click="() => to(n, index)()">
								{{ n.toText(userInfo) }}
							</a>
						</div>
						<span class="created" v-tooltip="formatDate(n.created)">
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
		</transition>
	</div>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

import NotificationService from '@/services/notification'
import User from '@/components/misc/user.vue'
import names from '@/models/constants/notificationNames.json'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {mapState} from 'vuex'

const LOAD_NOTIFICATIONS_INTERVAL = 10000

export default defineComponent({
	name: 'notifications',
	components: {User},
	data() {
		return {
			notificationService: new NotificationService(),
			allNotifications: [],
			showNotifications: false,
			interval: null,
		}
	},
	mounted() {
		this.loadNotifications()
		document.addEventListener('click', this.hidePopup)
		this.interval = setInterval(this.loadNotifications, LOAD_NOTIFICATIONS_INTERVAL)
	},
	beforeUnmount() {
		document.removeEventListener('click', this.hidePopup)
		clearInterval(this.interval)
	},
	computed: {
		unreadNotifications() {
			return this.notifications.filter(n => n.readAt === null).length
		},
		notifications() {
			return this.allNotifications ? this.allNotifications.filter(n => n.name !== '') : []
		},
		...mapState({
			userInfo: state => state.auth.info,
		}),
	},
	methods: {
		hidePopup(e) {
			if (this.showNotifications) {
				closeWhenClickedOutside(e, this.$refs.popup, () => this.showNotifications = false)
			}
		},
		async loadNotifications() {
			this.allNotifications = await this.notificationService.getAll()
		},
		to(n, index) {
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
					this.$router.push(to)
				}

				n.read = true
				this.allNotifications[index] = await this.notificationService.update(n)
			}
		},
	},
})
</script>

<style lang="scss" scoped>
.notifications {
	width: $navbar-icon-width;

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
		position: fixed;
		right: 1rem;
		margin-top: 1rem;
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