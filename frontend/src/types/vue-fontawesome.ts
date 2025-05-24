// copied and slightly modified from unmerged pull request that corrects types
// https://github.com/FortAwesome/vue-fontawesome/pull/355

import type { FaSymbol, FlipProp, IconLookup, IconProp, PullProp, SizeProp, Transform } from '@fortawesome/fontawesome-svg-core'
import type { DefineComponent } from 'vue'


interface FontAwesomeIconProps {
	border?: boolean
	fixedWidth?: boolean
	flip?: FlipProp
	icon: IconProp
	mask?: IconLookup
	listItem?: boolean
	pull?: PullProp
	pulse?: boolean
	rotation?: 90 | 180 | 270 | '90' | '180' | '270'
	swapOpacity?: boolean
	size?: SizeProp
	spin?: boolean
	transform?: Transform
	symbol?: FaSymbol
	title?: string | string[]
	inverse?: boolean
}

interface FontAwesomeLayersProps {
	fixedWidth?: boolean
}

interface FontAwesomeLayersTextProps {
	value: string | number
	transform?: object | string
	counter?: boolean
	position?: 'bottom-left' | 'bottom-right' | 'top-left' | 'top-right'
}

export type FontAwesomeIcon = DefineComponent<FontAwesomeIconProps>
export type FontAwesomeLayers = DefineComponent<FontAwesomeLayersProps>
export type FontAwesomeLayersText = DefineComponent<FontAwesomeLayersTextProps>
