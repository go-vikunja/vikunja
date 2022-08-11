import {string} from 'zod'

export const TextFieldSchema = string().transform((value) => value.trim()).default('')