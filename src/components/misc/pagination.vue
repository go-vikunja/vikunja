<template>
    <nav
        aria-label="pagination"
        class="pagination is-centered p-4"
        role="navigation"
        v-if="totalPages > 1"
    >
        <router-link
            :disabled="currentPage === 1"
            :to="getRouteForPagination(currentPage - 1)"
            class="pagination-previous"
            tag="button">
            {{ $t('misc.previous') }}
        </router-link>
        <router-link
            :disabled="currentPage === totalPages"
            :to="getRouteForPagination(currentPage + 1)"
            class="pagination-next"
            tag="button">
            {{ $t('misc.next') }}
        </router-link>
        <ul class="pagination-list">
            <template v-for="(p, i) in pages">
                <li :key="'page' + i" v-if="p.isEllipsis">
                    <span class="pagination-ellipsis">&hellip;</span>
                </li>
                <li :key="'page' + i" v-else>
                    <router-link
                        :aria-label="'Goto page ' + p.number"
                        :class="{ 'is-current': p.number === currentPage }"
                        :to="getRouteForPagination(p.number)"
                        class="pagination-link"
                    >
                        {{ p.number }}
                    </router-link>
                </li>
            </template>
        </ul>
    </nav>
</template>

<script>
function createPagination(totalPages, currentPage) {
	const pages = []
	for (let i = 0; i < totalPages; i++) {

		// Show ellipsis instead of all pages
		if (
			i > 0 && // Always at least the first page
			(i + 1) < totalPages && // And the last page
			(
				// And the current with current + 1 and current - 1
				(i + 1) > currentPage + 1 ||
				(i + 1) < currentPage - 1
			)
		) {
			// Only add an ellipsis if the last page isn't already one
			if (pages[i - 1] && !pages[i - 1].isEllipsis) {
				pages.push({
					number: 0,
					isEllipsis: true,
				})
			}
			continue
		}

		pages.push({
			number: i + 1,
			isEllipsis: false,
		})
	}
	return pages
}

function getRouteForPagination(page = 1, type = 'list') {
	return {
		name: 'list.' + type,
		params: {
			type: type,
		},
		query: {
			page: page,
		},
	}
}

export default {
	name: 'Pagination',

	props: {
		totalPages: {
			type: Number,
			required: true,
		},
		currentPage: {
			type: Number,
			required: true,
		},
	},

	computed: {
		pages() {
			return createPagination(this.totalPages, this.currentPage)
		},
	},

	methods: {
		getRouteForPagination,
	},
}
</script>

<style lang="scss" scoped>
.pagination {
  padding-bottom: 1rem;

  .pagination-previous,
  .pagination-next {
    &:not(:disabled):hover {
      background: $scheme-main;
      cursor: pointer;
    }
  }
}
</style>