<template>
	<notifications position="bottom left">
		<template slot="body" slot-scope="props">
			<div :class="['vue-notification-template', 'vue-notification', props.item.type]" @click="close(props)">
				<div
						v-if="props.item.title"
						class="notification-title"
						v-html="props.item.title"
				>
				</div>
				<div
						class="notification-content"
						v-html="props.item.text"
				>
				</div>
				<div class="buttons is-right" v-if="props.item.data && props.item.data.actions && props.item.data.actions.length > 0">
					<button
							class="button noshadow is-small"
							@click="action.callback"
							v-for="(action, i) in props.item.data.actions" :key="'action_'+i">
						{{ action.title }}
					</button>
				</div>
			</div>
		</template>
	</notifications>
</template>

<script>
	export default {
		name: 'notification',
		methods: {
			close(props) {
				props.close()
			},
		},
	}
</script>

<style scoped>

	.vue-notification {
		z-index: 9999;
	}

	.buttons {
		margin-top: .5em;
	}
</style>