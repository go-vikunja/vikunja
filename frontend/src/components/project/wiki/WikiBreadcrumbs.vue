<template>
	<nav
		class="breadcrumb"
		aria-label="breadcrumbs"
	>
		<ul>
			<li>
				<BaseButton
					@click="navigateToRoot"
				>
					<Icon icon="book" />
					<span>{{ $t('wiki.title') }}</span>
				</BaseButton>
			</li>
			<li
				v-for="(ancestor, index) in ancestors"
				:key="ancestor.id"
				:class="{ 'is-active': index === ancestors.length - 1 }"
			>
				<BaseButton
					v-if="index < ancestors.length - 1"
					@click="navigateToPage(ancestor)"
				>
					{{ ancestor.title }}
				</BaseButton>
				<span v-else>
					{{ ancestor.title }}
				</span>
			</li>
		</ul>
	</nav>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useRouter} from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'
import {useWikiPageStore} from '@/stores/wikiPages'
import type {IWikiPage} from '@/modelTypes/IWikiPage'

const props = defineProps<{
	projectId: number
	page: IWikiPage
}>()

const router = useRouter()
const wikiPageStore = useWikiPageStore()

const ancestors = computed(() => {
	return wikiPageStore.getAncestors(props.projectId, props.page)
})

function navigateToRoot() {
	router.push({
		query: {},
	})
}

function navigateToPage(page: IWikiPage) {
	if (page.isFolder) return
	
	router.push({
		query: {
			pageId: page.id,
		},
	})
}
</script>

<style lang="scss" scoped>
.breadcrumb {
	background: transparent;
	padding: 0.75rem 0;
	margin-bottom: 1rem;
	
	ul {
		flex-wrap: nowrap;
		display: flex;
		align-items: center;
		overflow-x: auto;
	}
	
	li {
		display: flex;
		align-items: center;
		
		+ li::before {
			content: '/';
			padding: 0 0.5rem;
			color: var(--grey-400);
		}
		
		&.is-active span {
			color: var(--text);
			font-weight: 600;
		}
		
		button {
			padding: 0.25rem 0.5rem;
			color: var(--primary);
			
			&:hover {
				text-decoration: underline;
			}
		}
	}
}
</style>
