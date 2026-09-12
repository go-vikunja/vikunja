<script setup lang="ts">
import {computed} from 'vue'
import {useLocalStorage} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import XButton from '@/components/input/Button.vue'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {MCP_CLIENT_HELP} from '@/urls'

const props = defineProps<{endpoint: string, token: string}>()
const {t} = useI18n({useScope: 'global'})
const copy = useCopyToClipboard()
const client = useLocalStorage('mcp-client', '')
const clients = Object.keys(MCP_CLIENT_HELP)
const helpUrl = computed(() => MCP_CLIENT_HELP[client.value])

function shellQuote(value: string) {
	return `'${value.replaceAll('\'', '\'\\\'\'')}'`
}

const instructions = computed<Record<string, {key: string, value?: string}[]>>(() => ({
	claudeCode: [
		{key: 'command', value: `claude mcp add --transport http vikunja ${shellQuote(props.endpoint)} --header ${shellQuote(`Authorization: Bearer ${props.token}`)}`},
		{key: 'verify'},
	],
	codex: [
		{key: 'environment', value: `export VIKUNJA_MCP_TOKEN=${shellQuote(props.token)}`},
		{key: 'command', value: `codex mcp add vikunja --url ${shellQuote(props.endpoint)} --bearer-token-env-var VIKUNJA_MCP_TOKEN`},
		{key: 'config', value: `[mcp_servers.vikunja]\nurl = ${JSON.stringify(props.endpoint)}\nbearer_token_env_var = "VIKUNJA_MCP_TOKEN"`},
		{key: 'note'},
	],
	claudeDesktop: [
		{key: 'open'},
		{key: 'url', value: props.endpoint},
		{key: 'auth'},
		{key: 'header', value: `Bearer ${props.token}`},
		{key: 'beta'},
	],
	mistral: [
		{key: 'open'},
		{key: 'url', value: props.endpoint},
		{key: 'token', value: props.token},
	],
	other: [
		{key: 'url', value: props.endpoint},
		{key: 'header', value: `Authorization: Bearer ${props.token}`},
	],
}))
const steps = computed(() => instructions.value[client.value] ?? [])
</script>

<template>
	<div class="mcp-client-guide">
		<div class="field">
			<label
				class="label"
				for="mcp-client"
			>{{ t('user.settings.mcp.client') }}</label>
			<div class="control select">
				<select
					id="mcp-client"
					v-model="client"
				>
					<option
						value=""
						disabled
					>
						{{ t('user.settings.mcp.chooseClient') }}
					</option>
					<option
						v-for="id in clients"
						:key="id"
						:value="id"
					>
						{{ t(`user.settings.mcp.clients.${id}.title`) }}
					</option>
				</select>
			</div>
		</div>
		<template v-if="client === 'chatgpt'">
			<p>{{ t('user.settings.mcp.clients.chatgpt.unavailable') }}</p>
		</template>
		<ol
			v-else-if="steps.length"
			class="mcp-steps"
		>
			<li
				v-for="step in steps"
				:key="step.key"
			>
				<i18n-t
					:keypath="`user.settings.mcp.clients.${client}.${step.key}`"
					tag="div"
					scope="global"
				>
					<template #value>
						<div
							v-if="step.value"
							class="mcp-copyable"
						>
							<pre><code>{{ step.value }}</code></pre>
							<XButton
								variant="secondary"
								type="button"
								@click="copy(step.value)"
							>
								{{ t('misc.copy') }}
							</XButton>
						</div>
					</template>
				</i18n-t>
			</li>
		</ol>
		<p v-if="helpUrl">
			<a
				:href="helpUrl"
				target="_blank"
				rel="noreferrer"
			>{{ client === 'other' ? t('user.settings.mcp.more') : t('user.settings.mcp.clientHelp', {client: t(`user.settings.mcp.clients.${client}.title`)}) }}</a>
		</p>
	</div>
</template>

<style scoped lang="scss">
.mcp-steps {
	margin-inline-start: 1.5rem;

	li {
		margin-block-end: 1rem;
	}
}

.mcp-copyable {
	margin-block: .5rem;

	pre {
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		margin-block-end: .5rem;
	}
}
</style>
