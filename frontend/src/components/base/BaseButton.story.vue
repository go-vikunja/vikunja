<script lang="ts" setup>
import {logEvent} from 'histoire/client'
import {reactive} from 'vue'
import {createRouter, createMemoryHistory} from 'vue-router'
import BaseButton from './BaseButton.vue'

function setupApp({ app }) {
	// Router mock
	app.use(createRouter({
		history: createMemoryHistory(),
		routes: [
			{ path: '/', name: 'home', component: { render: () => null } },
		],
	}))
}


const state = reactive({
	disabled: false,
})
</script>

<template>
	<Story
		:setup-app="setupApp"
		:layout="{ type: 'grid', width: '200px' }"
	>
		<Variant title="custom">
			<template #controls>
				<HstCheckbox
					v-model="state.disabled"
					title="Disabled"
				/>
			</template>
			<BaseButton :disabled="state.disabled">
				Hello!
			</BaseButton>
		</Variant>

		<Variant title="disabled">
			<BaseButton disabled>
				Hello!
			</BaseButton>
		</Variant>

		<Variant title="router link">
			<BaseButton :to="'home'">
				Hello!
			</BaseButton>
		</Variant>

		<Variant title="external link">
			<BaseButton href="https://vikunja.io">
				Hello!
			</BaseButton>
		</Variant>

		<Variant title="button">
			<BaseButton @click="logEvent('Click', $event)">
				Hello!
			</BaseButton>
		</Variant>
	</Story>
</template>
