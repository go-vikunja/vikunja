import {readonly, ref, computed} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'

import WikiPageService from '@/services/wikiPage'
import {setModuleLoading} from '@/stores/helper'

import type {IWikiPage} from '@/modelTypes/IWikiPage'

export const useWikiPageStore = defineStore('wikiPage', () => {
	const isLoading = ref(false)

	// Wiki pages stored by project ID
	const wikiPagesByProject = ref<{ [projectId: number]: { [pageId: number]: IWikiPage } }>({})

	const getWikiPagesForProject = computed(() => {
		return (projectId: number) => {
			const pages = wikiPagesByProject.value[projectId] || {}
			return Object.values(pages).sort((a, b) => {
				// Sort folders first, then by position
				if (a.isFolder !== b.isFolder) {
					return a.isFolder ? -1 : 1
				}
				return a.position - b.position
			})
		}
	})

	const getRootPagesForProject = computed(() => {
		return (projectId: number) => {
			return getWikiPagesForProject.value(projectId).filter(p => !p.parentId || p.parentId === 0)
		}
	})

	const getChildPages = computed(() => {
		return (projectId: number, parentId: number) => {
			return getWikiPagesForProject.value(projectId).filter(p => p.parentId === parentId)
		}
	})

	const getPageById = computed(() => {
		return (projectId: number, pageId: number) => {
			return wikiPagesByProject.value[projectId]?.[pageId] || null
		}
	})

	const getAncestors = computed(() => {
		return (projectId: number, page: IWikiPage): IWikiPage[] => {
			if (typeof page === 'undefined' || !page) {
				return []
			}

			if (!page.parentId || page.parentId === 0) {
				return [page]
			}

			const parentPage = getPageById.value(projectId, page.parentId)
			if (!parentPage) {
				return [page]
			}

			return [
				...getAncestors.value(projectId, parentPage),
				page,
			]
		}
	})

	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading
	}

	function setWikiPage(projectId: number, page: IWikiPage) {
		if (!wikiPagesByProject.value[projectId]) {
			wikiPagesByProject.value[projectId] = {}
		}
		wikiPagesByProject.value[projectId][page.id] = page
	}

	function setWikiPages(projectId: number, pages: IWikiPage[]) {
		pages.forEach(p => setWikiPage(projectId, p))
	}

	function removeWikiPage(projectId: number, pageId: number) {
		if (wikiPagesByProject.value[projectId]) {
			delete wikiPagesByProject.value[projectId][pageId]
		}
	}

	async function loadWikiPagesForProject(projectId: number) {
		const cancel = setModuleLoading(setIsLoading)
		const wikiPageService = new WikiPageService()

		try {
			// Fetch all pages with a high per_page limit to avoid pagination
			const pages = await wikiPageService.getAll({projectId}, {per_page: 9999}) as IWikiPage[]
			
			// Clear existing pages for this project
			wikiPagesByProject.value[projectId] = {}
			
			setWikiPages(projectId, pages)
			return pages
		} finally {
			cancel()
		}
	}

	async function loadWikiPage(projectId: number, pageId: number) {
		const cancel = setModuleLoading(setIsLoading)
		const wikiPageService = new WikiPageService()

		try {
			const page = await wikiPageService.get({projectId, id: pageId})
			setWikiPage(projectId, page)
			return page
		} finally {
			cancel()
		}
	}

	async function createWikiPage(projectId: number, page: IWikiPage) {
		const cancel = setModuleLoading(setIsLoading)
		const wikiPageService = new WikiPageService()

		try {
			// Ensure the page has the correct projectId
			const pageWithProjectId = {...page, projectId}
			const createdPage = await wikiPageService.create(pageWithProjectId)
			setWikiPage(projectId, createdPage)
			return createdPage
		} finally {
			cancel()
		}
	}

	async function updateWikiPage(projectId: number, page: IWikiPage) {
		const cancel = setModuleLoading(setIsLoading)
		const wikiPageService = new WikiPageService()

		try {
			// Ensure the page has the correct projectId
			const pageWithProjectId = {...page, projectId}
			const updatedPage = await wikiPageService.update(pageWithProjectId)
			setWikiPage(projectId, updatedPage)
			return updatedPage
		} finally {
			cancel()
		}
	}

	async function deleteWikiPage(projectId: number, page: IWikiPage) {
		const cancel = setModuleLoading(setIsLoading)
		const wikiPageService = new WikiPageService()

		try {
			// Ensure the page has the correct projectId
			const pageWithProjectId = {...page, projectId}
			await wikiPageService.delete(pageWithProjectId)
			removeWikiPage(projectId, page.id)
			
			// Remove child pages if it's a folder
			if (page.isFolder) {
				const childPages = getChildPages.value(projectId, page.id)
				childPages.forEach(child => removeWikiPage(projectId, child.id))
			}
		} finally {
			cancel()
		}
	}

	async function moveWikiPage(projectId: number, pageId: number, newParentId: number | null) {
		const cancel = setModuleLoading(setIsLoading)
		const wikiPageService = new WikiPageService()

		try {
			const movedPage = await wikiPageService.move(projectId, pageId, newParentId)
			setWikiPage(projectId, movedPage)
			return movedPage
		} finally {
			cancel()
		}
	}

	async function reorderWikiPage(projectId: number, pageId: number, position: number) {
		const cancel = setModuleLoading(setIsLoading)
		const wikiPageService = new WikiPageService()

		try {
			const reorderedPage = await wikiPageService.reorder(projectId, pageId, position)
			setWikiPage(projectId, reorderedPage)
			return reorderedPage
		} finally {
			cancel()
		}
	}

	async function searchWikiPages(projectId: number, query: string) {
		const cancel = setModuleLoading(setIsLoading)
		const wikiPageService = new WikiPageService()

		try {
			const results = await wikiPageService.search(projectId, query)
			return results
		} finally {
			cancel()
		}
	}

	return {
		isLoading: readonly(isLoading),
		wikiPagesByProject: readonly(wikiPagesByProject),

		getWikiPagesForProject,
		getRootPagesForProject,
		getChildPages,
		getPageById,
		getAncestors,

		setWikiPage,
		setWikiPages,
		removeWikiPage,

		loadWikiPagesForProject,
		loadWikiPage,
		createWikiPage,
		updateWikiPage,
		deleteWikiPage,
		moveWikiPage,
		reorderWikiPage,
		searchWikiPages,
	}
})

if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useWikiPageStore, import.meta.hot))
}
