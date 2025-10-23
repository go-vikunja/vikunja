import { z } from 'zod';
import { logError } from './logger.js';

/**
 * JSON-RPC 2.0 request schema
 * Used to validate incoming MCP protocol messages
 * 
 * References:
 * - JSON-RPC 2.0 Spec: https://www.jsonrpc.org/specification
 * - MCP Protocol uses JSON-RPC as transport layer
 */
export const JsonRpcRequestSchema = z.object({
	jsonrpc: z.literal('2.0'),
	method: z.string(),
	params: z.any().optional(),
	id: z.union([z.string(), z.number(), z.null()]).optional(),
});

/**
 * SSE transport POST /sse request body schema
 */
export const SSEMessageSchema = z.object({
	session_id: z.string().min(1),
	message: JsonRpcRequestSchema,
});

/**
 * Maximum request body size (1MB)
 * Prevents DoS attacks via large payloads
 */
export const MAX_REQUEST_BODY_SIZE = 1024 * 1024; // 1MB

/**
 * Validation error response format
 */
export interface ValidationErrorResponse {
	error: {
		code: number;
		message: string;
		data?: {
			validationErrors?: z.ZodIssue[];
		};
	};
}

/**
 * Validates a request body against a Zod schema
 * 
 * @param body - The request body to validate
 * @param schema - The Zod schema to validate against
 * @param maxSize - Maximum allowed body size in bytes
 * @returns Validation result with parsed data or error
 * 
 * @example
 * ```typescript
 * const result = validateRequestBody(req.body, JsonRpcRequestSchema);
 * if (!result.success) {
 *   res.status(400).json(result.error);
 *   return;
 * }
 * const validatedData = result.data;
 * ```
 */
export function validateRequestBody<T>(
	body: unknown,
	schema: z.ZodSchema<T>,
	maxSize: number = MAX_REQUEST_BODY_SIZE
): { success: true; data: T } | { success: false; error: ValidationErrorResponse } {
	try {
		// Check body size (approximate)
		const bodySize = JSON.stringify(body).length;
		if (bodySize > maxSize) {
			logError(new Error('Request body too large'), {
				event: 'request_validation_failed',
				bodySize,
				maxSize,
			});
			
			return {
				success: false,
				error: {
					error: {
						code: -32600,
						message: `Request body too large (${bodySize} bytes). Maximum allowed: ${maxSize} bytes`,
					},
				},
			};
		}

		// Validate schema
		const result = schema.safeParse(body);
		
		if (!result.success) {
			logError(new Error('Schema validation failed'), {
				event: 'request_validation_failed',
				errors: result.error.errors,
			});
			
			return {
				success: false,
				error: {
					error: {
						code: -32600,
						message: 'Invalid Request: Schema validation failed',
						data: {
							validationErrors: result.error.errors,
						},
					},
				},
			};
		}

		return {
			success: true,
			data: result.data,
		};
	} catch (error) {
		logError(error instanceof Error ? error : new Error(String(error)), {
			event: 'request_validation_error',
		});
		
		return {
			success: false,
			error: {
				error: {
					code: -32603,
					message: 'Internal error during request validation',
				},
			},
		};
	}
}

/**
 * Validates JSON-RPC request body
 * Convenience wrapper for common use case
 */
export function validateJsonRpcRequest(body: unknown) {
	return validateRequestBody(body, JsonRpcRequestSchema);
}

/**
 * Validates SSE message request body
 * Convenience wrapper for common use case
 */
export function validateSSEMessage(body: unknown) {
	return validateRequestBody(body, SSEMessageSchema);
}
