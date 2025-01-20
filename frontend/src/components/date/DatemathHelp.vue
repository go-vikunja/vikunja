<template>
	<Card
		class="has-no-shadow how-it-works-modal"
		:title="$t('input.datemathHelp.title')"
	>
		<p>
			{{ $t('input.datemathHelp.intro') }}
		</p>
		<p>
			<i18n-t
				keypath="input.datemathHelp.expression"
				scope="global"
			>
				<code>now</code>
				<code>||</code>
			</i18n-t>
		</p>
		<p>
			<i18n-t
				keypath="input.datemathHelp.similar"
				scope="global"
			>
				<BaseButton
					href="https://grafana.com/docs/grafana/latest/dashboards/time-range-controls/"
					target="_blank"
				>
					Grafana
				</BaseButton>
				<BaseButton
					href="https://www.elastic.co/guide/en/elasticsearch/reference/7.3/common-options.html#date-math"
					target="_blank"
				>
					Elasticsearch
				</BaseButton>
			</i18n-t>
		</p>
		<p>{{ $t('misc.forExample') }}</p>
		<ul>
			<li><code>+1d</code>: {{ $t('input.datemathHelp.add1Day') }}</li>
			<li><code>-1d</code>: {{ $t('input.datemathHelp.minus1Day') }}</li>
			<li><code>/d</code>: {{ $t('input.datemathHelp.roundDay') }}</li>
		</ul>
		<h3>{{ $t('input.datemathHelp.supportedUnits') }}</h3>
		<table class="table">
			<tbody>
				<tr>
					<td><code>s</code></td>
					<td>{{ $t('input.datemathHelp.units.seconds') }}</td>
				</tr>
				<tr>
					<td><code>m</code></td>
					<td>{{ $t('input.datemathHelp.units.minutes') }}</td>
				</tr>
				<tr>
					<td><code>h</code></td>
					<td>{{ $t('input.datemathHelp.units.hours') }}</td>
				</tr>
				<tr>
					<td><code>H</code></td>
					<td>{{ $t('input.datemathHelp.units.hours') }}</td>
				</tr>
				<tr>
					<td><code>d</code></td>
					<td>{{ $t('input.datemathHelp.units.days') }}</td>
				</tr>
				<tr>
					<td><code>w</code></td>
					<td>{{ $t('input.datemathHelp.units.weeks') }}</td>
				</tr>
				<tr>
					<td><code>M</code></td>
					<td>{{ $t('input.datemathHelp.units.months') }}</td>
				</tr>
				<tr>
					<td><code>y</code></td>
					<td>{{ $t('input.datemathHelp.units.years') }}</td>
				</tr>
			</tbody>
		</table>

		<h3>{{ $t('input.datemathHelp.someExamples') }}</h3>
		<table class="table">
			<tbody>
				<tr>
					<td><code>now</code></td>
					<td>{{ $t('input.datemathHelp.examples.now') }}</td>
				</tr>
				<tr>
					<td><code>now+24h</code></td>
					<td>{{ $t('input.datemathHelp.examples.in24h') }}</td>
				</tr>
				<tr>
					<td><code>now/d</code></td>
					<td>{{ $t('input.datemathHelp.examples.today') }}</td>
				</tr>
				<tr>
					<td><code>now/w</code></td>
					<td>{{ $t('input.datemathHelp.examples.beginningOfThisWeek') }}</td>
				</tr>
				<tr>
					<td><code>now/w+1w</code></td>
					<td>{{ $t('input.datemathHelp.examples.endOfThisWeek') }}</td>
				</tr>
				<tr>
					<td><code>now+30d</code></td>
					<td>{{ $t('input.datemathHelp.examples.in30Days') }}</td>
				</tr>
				<tr>
					<td><code>{{ exampleDate }}||+1M/d</code></td>
					<td>
						<i18n-t
							keypath="input.datemathHelp.examples.datePlusMonth"
							scope="global"
						>
							<strong>{{ exampleDate }}</strong>
						</i18n-t>
					</td>
				</tr>
			</tbody>
		</table>
	</Card>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import {useGlobalNow} from '@/composables/useGlobalNow'
import {formatDateShort} from '@/helpers/time/formatDate'

import BaseButton from '@/components/base/BaseButton.vue'

const {now} = useGlobalNow()
const exampleDate = computed(() => formatDateShort(now.value))
</script>

<style scoped lang="scss">
// FIXME: Remove style overwrites
.how-it-works-modal {
	font-size: 1rem;
}

p {
	display: inline-block !important;
}

.base-button {
	display: inline;
}
</style>
