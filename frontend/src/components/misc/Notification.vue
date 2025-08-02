<template>
	<Notifications
		position="bottom left"
		:max="2"
		class="global-notification"
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
					{{ item.title }}
				</div>
				<div class="notification-content">
					<template
						v-for="(t, k) in item.text"
						:key="k"
					>
						{{ t }}<br>
					</template>
				</div>
				<div
					v-if="item.data?.actions?.length > 0"
					class="buttons is-right"
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
</template>

<style scoped>
.vue-notification {
	z-index: 9999;
}

.buttons {
	margin-block-start: 0.5rem;
}
</style>
