import type {TypeOf} from 'zod'
import {string} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'
import {HexColorSchema} from './common/hexColor'

import {UserSchema} from './user'
import {AbstractSchema} from './abstract'

import {colorIsDark} from '@/helpers/color/colorIsDark'

const DEFAULT_LABEL_BACKGROUND_COLOR = 'e8e8e8'

export const LabelSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	title: string().default(''),
	hexColor: HexColorSchema.default(DEFAULT_LABEL_BACKGROUND_COLOR),
	textColor: string(), // implicit
	description: string().default(''),
	createdBy: UserSchema, // FIXME: default: current user?
	listId: IdSchema.default(0),

	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
}).transform((obj) => {
	// FIXME: remove textColor location => should be defined in UI
	obj.textColor = colorIsDark(obj.hexColor) ? '#4a4a4a' : '#ffffff'
	return obj
},
)

export type ILabel = TypeOf<typeof LabelSchema> 
