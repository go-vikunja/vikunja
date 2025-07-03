/**
 * Common type definitions for service layer
 * These interfaces replace `any` types throughout the service layer
 */

// Generic API Response Types
export interface ApiResponse<T = unknown> {
	data: T
	headers: Record<string, string>
	status: number
}

export interface PaginatedResponse<T> {
	data: T[]
	meta: {
		currentPage: number
		totalPages: number
		totalItems: number
		itemsPerPage: number
	}
}

export interface BlobResponse {
	data: Blob
	type: string
	size: number
}

// Service Method Parameters
export interface RouteParams {
	[key: string]: string | number | undefined
}

export interface QueryParams {
	page?: number
	per_page?: number
	sort_by?: string[]
	order_by?: ('asc' | 'desc')[]
	filter?: string
	search?: string
	[key: string]: unknown
}

export interface FileUploadParams {
	files: File[] | FileList
	fieldName: string
	additionalData?: Record<string, unknown>
}

// Service Factory Response Types
export interface ModelFactoryResponse<T> {
	success: T[] | null
	errors?: string[]
}

export interface AttachmentUploadResponse {
	success: IAttachment[] | null
	errors?: string[]
}

export interface BackgroundImageResponse {
	url: string
	blurHash?: string
	info?: {
		width: number
		height: number
		author?: string
		source?: string
	}
}

// Authentication & Authorization Types
export interface ApiRoutesResponse {
	routes: {
		[method: string]: string[]
	}
}

export interface WebhookEventsResponse {
	events: string[]
}

export interface PasswordResetRequest {
	email: string
}

export interface PasswordResetResponse {
	message: string
	token?: string
}

// Data Export Types
export interface DataExportRequest {
	password: string
}

export interface DataExportResponse {
	message: string
	downloadUrl?: string
	status: 'pending' | 'processing' | 'completed' | 'failed'
}

// Generic Service Configuration
export interface ServiceConfig {
	baseURL?: string
	timeout?: number
	headers?: Record<string, string>
	retryAttempts?: number
}

export interface ServiceMethods {
	get: string
	getAll: string
	create: string
	update: string
	delete: string
	reset?: string
	[key: string]: string | undefined
}

// Model Processing Types
export type ModelProcessor<T> = (model: T) => T

export interface DateProcessingResult {
	iso: string
	timestamp: number
	formatted: string
}

export interface ValidationResult {
	isValid: boolean
	errors: string[]
	warnings?: string[]
}

// HTTP Request Types
export interface HttpHeaders {
	[key: string]: string
}

export interface RequestConfig {
	method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
	headers?: HttpHeaders
	params?: QueryParams
	data?: unknown
	timeout?: number
}

// Error Response Types
export interface ErrorResponse {
	error: string
	message: string
	statusCode: number
	details?: Record<string, unknown>
}

// Generic CRUD Operations
export interface CrudOperations<T> {
	getAll: (params?: QueryParams) => Promise<T[]>
	get: (id: number | string) => Promise<T>
	create: (data: Partial<T>) => Promise<T>
	update: (data: T) => Promise<T>
	delete: (id: number | string) => Promise<void>
}

// Import IAttachment type for AttachmentUploadResponse
import type { IAttachment } from '@/modelTypes/IAttachment'