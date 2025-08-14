<template>
	<div class="no-auth-wrapper">
		<Logo
			class="logo"
			width="200"
			height="58"
		/>
		<div class="noauth-container">
			<section
				class="image"
				:class="{ 'has-message': motd !== '' }"
			>
				<Message v-if="motd !== ''">
					{{ motd }}
				</Message>
				<h2 class="image-title">
					{{ $t("misc.welcomeBack") }}
				</h2>
			</section>
			<section class="content">
				<div>
					<h2
						v-if="title"
						class="title"
					>
						{{ title }}
					</h2>
					<ApiConfig v-if="showApiConfig" />
					<Message
						v-if="motd !== ''"
						class="is-hidden-tablet mbe-4"
					>
						{{ motd }}
					</Message>
					<slot />
				</div>
				<Legal />
			</section>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

import Logo from '@/components/home/Logo.vue'
import Message from '@/components/misc/Message.vue'
import Legal from '@/components/misc/Legal.vue'
import ApiConfig from '@/components/misc/ApiConfig.vue'

import { useTitle } from '@/composables/useTitle'
import { useConfigStore } from '@/stores/config'

withDefaults(
	defineProps<{
		showApiConfig?: boolean;
	}>(),
	{
		showApiConfig: false,
	},
)
const configStore = useConfigStore()
const motd = computed(() => configStore.motd)

const route = useRoute()
const { t } = useI18n({ useScope: 'global' })
const title = computed(() =>
	route.meta?.title ? t(route.meta.title as string) : '',
)
useTitle(() => title.value)
</script>

<style lang="scss" scoped>
.no-auth-wrapper {
	background: var(--site-background) url("@/assets/llama.svg?url") no-repeat
		fixed bottom left;
	min-block-size: 100vh;
	display: flex;
	flex-direction: column;
	place-items: center;

	@media screen and (max-width: $fullhd) {
		padding-block-end: 15rem;
	}
}

.noauth-container {
	max-inline-size: $desktop;
	inline-size: 100%;
	min-block-size: 60vh;
	display: flex;
	background-color: var(--white);
	box-shadow: var(--shadow-md);
	overflow: hidden;

	@media screen and (min-width: $desktop) {
		border-radius: $radius;
	}
}

.image {
	inline-size: 50%;
	padding: 1rem;
	display: flex;
	flex-direction: column;
	justify-content: flex-end;

	@media screen and (max-width: $tablet) {
		display: none;
	}

	@media screen and (min-width: $tablet) {
		background: url("@/assets/no-auth-image.jpg") no-repeat bottom/cover;
		position: relative;

		&.has-message {
			justify-content: space-between;
		}

		&::before {
			content: "";
			position: absolute;
			inset-block-start: 0;
			inset-inline-start: 0;
			inset-inline-end: 0;
			inset-block-end: 0;
			background-color: rgba(0, 0, 0, 0.2);
		}

		> * {
			position: relative;
		}
	}
}

.content {
	display: flex;
	justify-content: space-between;
	flex-direction: column;
	padding: 2rem 2rem 1.5rem;

	@media screen and (max-width: $desktop) {
		inline-size: 100%;
		max-inline-size: 450px;
		margin-inline: auto;
	}

	@media screen and (min-width: $desktop) {
		inline-size: 50%;
	}
}

.logo {
	max-inline-size: 100%;
	margin: 1rem 0;
}

.image-title {
	color: hsl(0deg, 0%, 100%);
	font-size: 2.5rem;
}
</style>
