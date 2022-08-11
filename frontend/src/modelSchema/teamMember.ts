import type {TypeOf} from 'zod'
import {boolean} from 'zod'

import {IdSchema} from './common/id'

import {UserSchema} from './user'

export const TeamMemberSchema = UserSchema.extend({
	admin: boolean().default(false),
	teamId: IdSchema.default(0), // IList['id']
})

export type TeamMember = TypeOf<typeof TeamMemberSchema> 