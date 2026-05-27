<template>
	<nav
		v-if="parentChain.length > 0"
		class="project-breadcrumb"
		aria-label="Project hierarchy"
	>
		<ol class="breadcrumb-list">
			<li
				v-for="(project, index) in parentChain"
				:key="project.id"
				class="breadcrumb-item"
			>
				<RouterLink
					:to="{name: 'project.index', params: {projectId: project.id}}"
					class="breadcrumb-link"
				>
					{{ project.title }}
				</RouterLink>
				<span
					v-if="index < parentChain.length - 1"
					class="breadcrumb-separator"
					aria-hidden="true"
				>›</span>
			</li>
		</ol>
	</nav>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useProjectStore} from '@/stores/projects'
import type {IProject} from '@/modelTypes/IProject'

const props = defineProps<{
	projectId: IProject['id']
}>()

const projectStore = useProjectStore()

/**
 * Builds the ancestor chain from root down to (but not including) the current project.
 * Uses the already-loaded project store — no extra API calls needed.
 */
const parentChain = computed<IProject[]>(() => {
	const chain: IProject[] = []
	const current = projectStore.projects[props.projectId] as IProject | undefined

	if (!current || !current.parentProjectId) {
		return chain
	}

	// Walk up the parent chain, collecting ancestors
	const visited = new Set<number>()
	let parentId = current.parentProjectId

	while (parentId && !visited.has(parentId)) {
		visited.add(parentId)
		const parent = projectStore.projects[parentId] as IProject | undefined
		if (!parent) break
		chain.unshift(parent) // prepend so root comes first
		parentId = parent.parentProjectId
	}

	return chain
})
</script>

<style lang="scss" scoped>
.project-breadcrumb {
	font-size: .8rem;
	color: var(--grey-500);
	margin-block-end: .25rem;
}

.breadcrumb-list {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .25rem;
	list-style: none;
	margin: 0;
	padding: 0;
}

.breadcrumb-item {
	display: flex;
	align-items: center;
	gap: .25rem;
}

.breadcrumb-link {
	color: var(--grey-500);
	text-decoration: none;
	transition: color 100ms;

	&:hover {
		color: var(--primary);
	}
}

.breadcrumb-separator {
	color: var(--grey-400);
}
</style>
