<template>
	<div class="notifications">
		<a @click.stop="showNotifications = !showNotifications" class="trigger-button">
			<span class="unread-indicator" v-if="unreadNotifications > 0"></span>
			<icon icon="bell"/>
		</a>

		<transition name="fade">
			<div class="notifications-list" v-if="showNotifications" ref="popup">
				<span class="head">Notifications</span>
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
					<span class="detail">
						<p>

						<span class="has-text-weight-bold" v-if="n.notification.doer">
							{{ n.notification.doer.getDisplayName() }}
						</span>
						<a @click="() => to(n, index)()">
							{{ n.toText(userInfo) }}
						</a>
						</p>
					<div class="created" v-tooltip="formatDate(n.created)">
						{{ formatDateSince(n.created) }}
					</div>
					</span>
				</div>
				<p class="nothing" v-if="notifications.length === 0">
					You don't have any notifications. Have a nice day!<br/>
					<span class="explainer">
						Notifications will appear here when actions on namespaces, lists or tasks you subscribed to happen.
					</span>
				</p>
			</div>
		</transition>
	</div>
</template>

<script>
import NotificationService from '@/services/notification'
import User from '@/components/misc/user'
import names from '@/models/notificationNames.json'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {mapState} from 'vuex'

export default {
	name: 'notifications',
	components: {User},
	data() {
		return {
			notificationService: NotificationService,
			allNotifications: [],
			showNotifications: false,
			interval: null,
		}
	},
	created() {
		this.notificationService = new NotificationService()
	},
	mounted() {
		this.loadNotifications()
		document.addEventListener('click', this.hidePopup)
		this.interval = setInterval(this.loadNotifications, 10000)
	},
	beforeDestroy() {
		document.removeEventListener('click', this.hidePopup)
		clearInterval(this.interval)
	},
	computed: {
		unreadNotifications() {
			return this.notifications.filter(n => n.readAt === null).length
		},
		notifications() {
			return this.allNotifications.filter(n => n.name !== '')
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
		loadNotifications() {
			this.notificationService.getAll()
				.then(r => {
					this.$set(this, 'allNotifications', r)
				})
				.catch(e => {
					this.error(e, this)
				})
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

			return () => {
				if (to.name !== '') {
					this.$router.push(to)
				}

				n.read = true
				this.notificationService.update(n)
					.then(r => {
						this.$set(this.allNotifications, index, r)
					})
					.catch(e => this.error(e, this))
			}
		},
	},
}
</script>
