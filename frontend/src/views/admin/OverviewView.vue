<template>
	<Card :title="$t('navigation.overview')">
		<div class="admin-overview">
			<p v-if="loading">
				{{ $t('misc.loading') }}
			</p>
			<div
				v-else-if="data"
				class="admin-overview__grid"
			>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.labels.users') }}
					</h2>
					<p class="admin-overview__card-value">
						{{ data.users }}
					</p>
				</div>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('project.projects') }}
					</h2>
					<p class="admin-overview__card-value">
						{{ data.projects }}
					</p>
				</div>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.labels.tasks') }}
					</h2>
					<p class="admin-overview__card-value">
						{{ data.tasks }}
					</p>
				</div>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('team.title') }}
					</h2>
					<p class="admin-overview__card-value">
						{{ data.teams }}
					</p>
				</div>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.overview.shares') }}
					</h2>
					<p class="admin-overview__card-value">
						{{ totalShares }}
					</p>
					<p class="admin-overview__hint admin-overview__shares-breakdown">
						{{ data.shares.linkShares }} {{ $t('admin.overview.linkSharesShort') }}
						<span aria-hidden="true">·</span>
						{{ data.shares.teamShares }} {{ $t('admin.overview.teamSharesShort') }}
						<span aria-hidden="true">·</span>
						{{ data.shares.userShares }} {{ $t('admin.overview.userSharesShort') }}
					</p>
				</div>
				<div class="admin-overview__card admin-overview__card--version">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.overview.version') }}
					</h2>
					<p class="admin-overview__card-value admin-overview__card-value--version">
						{{ data.version }}
					</p>
				</div>
				<div class="admin-overview__card admin-overview__card--wide">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.overview.license') }}
					</h2>
					<dl class="admin-overview__kv">
						<dt>{{ $t('admin.overview.licenseValidUntil') }}</dt>
						<dd>
							<TimeDisplay :date="data.license.expiresAt" />
							<span
								v-if="expiresInDays !== null"
								class="admin-overview__hint"
							>
								({{ $t('admin.overview.licenseExpiresIn', {days: expiresInDays}) }})
							</span>
						</dd>
						<dt>{{ $t('admin.overview.licenseLastVerified') }}</dt>
						<dd>
							<TimeDisplay
								:date="data.license.validatedAt"
								mode="relative"
								:fallback="$t('admin.overview.licenseNever')"
							/>
							<span
								v-if="data.license.lastCheckFailed"
								class="has-text-danger admin-overview__hint"
							>
								({{ $t('admin.overview.licenseLastCheckFailed') }})
							</span>
						</dd>
						<template v-if="data.license.features.length">
							<dt>{{ $t('admin.overview.licenseFeatures') }}</dt>
							<dd>{{ data.license.features.join(', ') }}</dd>
						</template>
						<template v-if="data.license.instanceId">
							<dt>{{ $t('admin.overview.licenseInstance') }}</dt>
							<dd><code>{{ data.license.instanceId }}</code></dd>
						</template>
					</dl>
					<p class="admin-overview__card-action">
						<a
							href="https://console.vikunja.io"
							target="_blank"
							rel="noopener noreferrer"
						>
							{{ $t('admin.overview.licenseManage') }}
							<Icon
								icon="arrow-up-right-from-square"
								class="admin-overview__external-icon"
							/>
						</a>
					</p>
				</div>
			</div>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, computed, onMounted} from 'vue'
import dayjs from 'dayjs'
import Card from '@/components/misc/Card.vue'
import Icon from '@/components/misc/Icon'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'
import AdminOverviewService from '@/services/admin/overviewService'
import type {IAdminOverview} from '@/modelTypes/IAdminOverview'
import {error} from '@/message'

const adminOverviewService = new AdminOverviewService()

const data = ref<IAdminOverview | null>(null)
const loading = ref(false)

const expiresInDays = computed<number | null>(() => {
	const expiresAt = data.value?.license?.expiresAt
	if (!expiresAt) return null
	return Math.max(0, dayjs(expiresAt).diff(dayjs(), 'day'))
})

const totalShares = computed<number>(() => {
	const shares = data.value?.shares
	if (!shares) return 0
	return shares.linkShares + shares.teamShares + shares.userShares
})

onMounted(async () => {
	loading.value = true
	try {
		data.value = await adminOverviewService.getOverview()
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
})
</script>

<style lang="scss" scoped>
.admin-overview__grid {
	display: grid;
	grid-template-columns: 1fr;
	gap: 1rem;
}

@media screen and (min-width: $tablet) {
	.admin-overview__grid {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}
}

@media screen and (min-width: $desktop) {
	.admin-overview__grid {
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}
}

.admin-overview__card {
	position: relative;
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: 6px;
	padding: 1.25rem;
}

.admin-overview__card:not(.admin-overview__card--wide) {
	min-block-size: 7.5rem;
}

.admin-overview__card--version {
	container-type: inline-size;
}

.admin-overview__card-title {
	font-size: 0.85rem;
	color: var(--grey-600);
	text-transform: uppercase;
	letter-spacing: 0.03em;
	margin-block-end: 0.5rem;
}

.admin-overview__card-value {
	font-size: 1.75rem;
	font-weight: 600;
}

.admin-overview__card-value--version {
	font-size: clamp(0.9rem, 8cqi, 1.75rem);
	word-break: break-all;
	overflow-wrap: anywhere;
}

.admin-overview__card--wide {
	grid-column: 1 / -1;
}

.admin-overview__kv {
	display: grid;
	grid-template-columns: max-content 1fr;
	column-gap: 1rem;
	row-gap: 0.25rem;
	margin-block-start: 1rem;
	font-size: 0.9rem;

	dt {
		font-weight: 600;
		color: var(--grey-700);
	}

	dd {
		margin: 0;
	}
}

.admin-overview__hint {
	color: var(--grey-600);
	margin-inline-start: 0.25rem;
}

.admin-overview__card-action {
	margin-block-start: 1rem;
}

.admin-overview__shares-breakdown {
	position: absolute;
	inset-block-end: 1.25rem;
	inset-inline: 1.25rem;
	margin: 0;
	font-size: 0.75rem;
	color: var(--grey-500);
	text-align: end;
}

.admin-overview__external-icon {
	margin-inline-start: 0.35em;
	font-size: 0.85em;
	opacity: 0.7;
}
</style>
