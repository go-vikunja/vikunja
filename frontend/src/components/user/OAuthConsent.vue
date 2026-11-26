<template>
	<Card
		class="oauth-consent"
		:title="$t('user.auth.oauthConsent.title', {client: clientLabel})"
	>
		<p>
			<i18n-t
				keypath="user.auth.oauthConsent.intro"
				scope="global"
			>
				<template #client>
					<strong class="client">{{ clientLabel }}</strong>
				</template>
			</i18n-t>
		</p>

		<div class="account">
			<span class="account-label">{{ $t('user.auth.oauthConsent.account') }}</span>
			<strong>{{ userLabel }}</strong>
		</div>

		<fieldset
			v-if="requested === 'write'"
			class="access"
		>
			<legend class="access-legend">
				{{ $t('user.auth.oauthConsent.access') }}
			</legend>
			<label
				v-for="option in ACCESS_OPTIONS"
				:key="option.level"
				class="access-option"
				:class="{'is-selected': level === option.level}"
			>
				<input
					v-model="level"
					type="radio"
					name="oauthConsentAccess"
					:value="option.level"
				>
				<span>
					<strong>{{ $t(option.label) }}</strong>
					<span class="access-description">{{ $t(option.description) }}</span>
				</span>
			</label>
		</fieldset>
		<div
			v-else
			class="access"
		>
			<strong>{{ $t('project.share.permission.read') }}</strong>
			<span class="access-description">{{ $t('user.auth.oauthConsent.readDescription') }}</span>
		</div>

		<p class="details">
			{{ $t('user.auth.oauthConsent.duration') }}
			<i18n-t
				keypath="user.auth.oauthConsent.redirect"
				scope="global"
			>
				<template #host>
					<strong class="host">{{ redirectHost }}</strong>
				</template>
			</i18n-t>
		</p>

		<div class="buttons">
			<XButton
				:loading="pending"
				@click="approve"
			>
				{{ $t('user.auth.oauthConsent.approve') }}
			</XButton>
			<XButton
				variant="tertiary"
				@click="deny"
			>
				{{ $t('user.auth.oauthConsent.deny') }}
			</XButton>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'

import Card from '@/components/misc/Card.vue'
import {getDisplayName} from '@/helpers/user'
import type {
	AuthorizeClient,
	User as IUser,
} from '@/client/generated'

export type AccessLevel = 'read' | 'write'

const props = defineProps<{
	client: AuthorizeClient,
	redirectUri: string,
	requested: AccessLevel,
	user: IUser,
	pending: boolean,
}>()

const emit = defineEmits<{
	approve: [level: AccessLevel],
	deny: [],
}>()

const ACCESS_OPTIONS = [
	{
		level: 'write',
		label: 'project.share.permission.readWrite',
		description: 'user.auth.oauthConsent.writeDescription',
	},
	{
		level: 'read',
		label: 'project.share.permission.read',
		description: 'user.auth.oauthConsent.readDescription',
	},
] as const satisfies readonly {level: AccessLevel, label: string, description: string}[]

// eslint-disable-next-line vue/no-setup-props-reactivity-loss
const level = ref<AccessLevel>(props.requested)

const clientLabel = computed(() => props.client.id ?? '')
const userLabel = computed(() => {
	const name = getDisplayName(props.user)
	return name === props.user.username ? name : `${name} (${props.user.username})`
})
const redirectHost = computed(() => new URL(props.redirectUri).host)

function approve() {
	if (props.pending) return
	emit('approve', level.value)
}

function deny() {
	if (props.pending) return
	emit('deny')
}
</script>

<style lang="scss" scoped>
.oauth-consent {
	max-inline-size: 36rem;
	margin-inline: auto;
}

.client,
.host {
	overflow-wrap: anywhere;
}

.account {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .25rem;
	margin-block: 1rem;
}

.account-label {
	color: var(--grey-500);
}

.access {
	display: flex;
	flex-direction: column;
	gap: .5rem;
	margin-block: 1rem;
	padding: 0;
	border: 0;
}

.access-legend {
	font-weight: bold;
	margin-block-end: .5rem;
}

.access-option {
	display: flex;
	align-items: flex-start;
	gap: .75rem;
	padding: .75rem;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	cursor: pointer;

	&.is-selected {
		border-color: var(--primary);
	}

	input {
		margin-block-start: .25rem;
		cursor: pointer;
	}
}

.access-description {
	display: block;
	color: var(--grey-500);
}

.details {
	color: var(--grey-500);
	font-size: .875rem;
}

.buttons {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem;
}
</style>
