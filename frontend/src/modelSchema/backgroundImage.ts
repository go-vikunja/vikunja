import type {TypeOf} from 'zod'
import {object, record, string, unknown} from 'zod'

import {IdSchema} from './common/id'

export const BackgroundImageSchema = object({
  id: IdSchema.default(0),
  url: string().url().default(''),
  thumb: string().default(''),
	// FIXME: not sure if this needs to defined, since it seems provider specific
	// {
  //   author: string(),
  //   authorName: string(),
  // }
  info: record(unknown()).default({}),
  blurHash: string().default(''),
})

export type BackgroundImage = TypeOf<typeof BackgroundImageSchema> 