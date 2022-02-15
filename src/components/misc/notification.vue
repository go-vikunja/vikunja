<template>
	<notifications position="bottom left" :max="2" class="global-notification">
		<template #body="{ item, close }">
			<div
				:class="[
					'vue-notification-template',
					'vue-notification',
					item.type,
				]"
				@click="close()"
			>
				<div v-if="item.title" class="notification-title">{{ item.title }}</div>
				<div class="notification-content">
					<template v-for="(t, k) in item.text" :key="k">{{ t }}<br /></template>
				</div>
				<div
					class="buttons is-right"
					v-if="
						item.data &&
						item.data.actions &&
						item.data.actions.length > 0
					"
				>
					<x-button
						:key="'action_' + i"
						@click="action.callback"
						:shadow="false"
						class="is-small"
						variant="secondary"
						v-for="(action, i) in item.data.actions"
					>
						{{ action.title }}
					</x-button>
				</div>
			</div>
		</template>
	</notifications>
</template>

<script lang="ts">
export default {
	name: 'notification',
}
</script>

<style scoped>
.vue-notification {
	z-index: 9999;
}

.buttons {
	margin-top: 0.5rem;
}
</style>