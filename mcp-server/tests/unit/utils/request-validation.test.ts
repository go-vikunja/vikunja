import { describe, it, expect } from 'vitest';
import {
	validateJsonRpcRequest,
	validateSSEMessage,
	validateRequestBody,
	JsonRpcRequestSchema,
	SSEMessageSchema,
	MAX_REQUEST_BODY_SIZE,
} from '../../../src/utils/request-validation.js';

describe('Request Validation', () => {
	describe('validateJsonRpcRequest', () => {
		it('should validate valid JSON-RPC 2.0 request', () => {
			const validRequest = {
				jsonrpc: '2.0',
				method: 'tools/list',
				id: 1,
			};

			const result = validateJsonRpcRequest(validRequest);
			expect(result.success).toBe(true);
			if (result.success) {
				expect(result.data).toEqual(validRequest);
			}
		});

		it('should validate JSON-RPC request with params', () => {
			const validRequest = {
				jsonrpc: '2.0',
				method: 'tools/call',
				params: { name: 'get_tasks', arguments: {} },
				id: 2,
			};

			const result = validateJsonRpcRequest(validRequest);
			expect(result.success).toBe(true);
			if (result.success) {
				expect(result.data).toEqual(validRequest);
			}
		});

		it('should validate JSON-RPC notification (no id)', () => {
			const notification = {
				jsonrpc: '2.0',
				method: 'notifications/initialized',
			};

			const result = validateJsonRpcRequest(notification);
			expect(result.success).toBe(true);
		});

		it('should reject request without jsonrpc field', () => {
			const invalidRequest = {
				method: 'tools/list',
			};

			const result = validateJsonRpcRequest(invalidRequest);
			expect(result.success).toBe(false);
			if (!result.success) {
				expect(result.error.error.code).toBe(-32600);
				expect(result.error.error.message).toContain('Schema validation failed');
			}
		});

		it('should reject request with wrong jsonrpc version', () => {
			const invalidRequest = {
				jsonrpc: '1.0',
				method: 'tools/list',
			};

			const result = validateJsonRpcRequest(invalidRequest);
			expect(result.success).toBe(false);
			if (!result.success) {
				expect(result.error.error.code).toBe(-32600);
			}
		});

		it('should reject request without method field', () => {
			const invalidRequest = {
				jsonrpc: '2.0',
				id: 1,
			};

			const result = validateJsonRpcRequest(invalidRequest);
			expect(result.success).toBe(false);
			if (!result.success) {
				expect(result.error.error.code).toBe(-32600);
			}
		});

		it('should accept null id (per JSON-RPC spec)', () => {
			const validRequest = {
				jsonrpc: '2.0',
				method: 'tools/list',
				id: null,
			};

			const result = validateJsonRpcRequest(validRequest);
			expect(result.success).toBe(true);
		});

		it('should accept string id', () => {
			const validRequest = {
				jsonrpc: '2.0',
				method: 'tools/list',
				id: 'request-123',
			};

			const result = validateJsonRpcRequest(validRequest);
			expect(result.success).toBe(true);
		});
	});

	describe('validateSSEMessage', () => {
		it('should validate valid SSE message', () => {
			const validMessage = {
				session_id: 'session-123',
				message: {
					jsonrpc: '2.0',
					method: 'tools/list',
					id: 1,
				},
			};

			const result = validateSSEMessage(validMessage);
			expect(result.success).toBe(true);
			if (result.success) {
				expect(result.data).toEqual(validMessage);
			}
		});

		it('should reject message without session_id', () => {
			const invalidMessage = {
				message: {
					jsonrpc: '2.0',
					method: 'tools/list',
					id: 1,
				},
			};

			const result = validateSSEMessage(invalidMessage);
			expect(result.success).toBe(false);
			if (!result.success) {
				expect(result.error.error.code).toBe(-32600);
			}
		});

		it('should reject message with empty session_id', () => {
			const invalidMessage = {
				session_id: '',
				message: {
					jsonrpc: '2.0',
					method: 'tools/list',
					id: 1,
				},
			};

			const result = validateSSEMessage(invalidMessage);
			expect(result.success).toBe(false);
		});

		it('should reject message with invalid JSON-RPC message', () => {
			const invalidMessage = {
				session_id: 'session-123',
				message: {
					// Missing jsonrpc field
					method: 'tools/list',
					id: 1,
				},
			};

			const result = validateSSEMessage(invalidMessage);
			expect(result.success).toBe(false);
		});

		it('should reject message without message field', () => {
			const invalidMessage = {
				session_id: 'session-123',
			};

			const result = validateSSEMessage(invalidMessage);
			expect(result.success).toBe(false);
		});
	});

	describe('Body Size Validation', () => {
		it('should reject oversized request body', () => {
			// Create a large object exceeding MAX_REQUEST_BODY_SIZE
			const largeBody = {
				jsonrpc: '2.0',
				method: 'tools/call',
				params: {
					data: 'x'.repeat(MAX_REQUEST_BODY_SIZE + 1000),
				},
				id: 1,
			};

			const result = validateRequestBody(largeBody, JsonRpcRequestSchema);
			expect(result.success).toBe(false);
			if (!result.success) {
				expect(result.error.error.code).toBe(-32600);
				expect(result.error.error.message).toContain('Request body too large');
			}
		});

		it('should accept request body at size limit', () => {
			// Create object just under the limit
			const largeButValidBody = {
				jsonrpc: '2.0',
				method: 'tools/call',
				params: {
					data: 'x'.repeat(MAX_REQUEST_BODY_SIZE - 100),
				},
				id: 1,
			};

			const result = validateRequestBody(largeButValidBody, JsonRpcRequestSchema);
			expect(result.success).toBe(true);
		});

		it('should use custom max size when provided', () => {
			const smallLimit = 100; // 100 bytes
			const body = {
				jsonrpc: '2.0',
				method: 'tools/call',
				params: { data: 'x'.repeat(200) },
				id: 1,
			};

			const result = validateRequestBody(body, JsonRpcRequestSchema, smallLimit);
			expect(result.success).toBe(false);
			if (!result.success) {
				expect(result.error.error.message).toContain('Request body too large');
			}
		});
	});

	describe('Malformed Input', () => {
		it('should handle null input', () => {
			const result = validateJsonRpcRequest(null);
			expect(result.success).toBe(false);
			if (!result.success) {
				expect(result.error.error.code).toBe(-32600);
			}
		});

		it('should handle undefined input', () => {
			const result = validateJsonRpcRequest(undefined);
			expect(result.success).toBe(false);
		});

		it('should handle non-object input', () => {
			const result = validateJsonRpcRequest('not an object');
			expect(result.success).toBe(false);
		});

		it('should handle array input', () => {
			const result = validateJsonRpcRequest([{ jsonrpc: '2.0', method: 'test' }]);
			expect(result.success).toBe(false);
		});

		it('should handle number input', () => {
			const result = validateJsonRpcRequest(12345);
			expect(result.success).toBe(false);
		});

		it('should handle boolean input', () => {
			const result = validateJsonRpcRequest(true);
			expect(result.success).toBe(false);
		});
	});

	describe('Injection Attack Prevention', () => {
		it('should sanitize and validate request with special characters', () => {
			const potentialInjection = {
				jsonrpc: '2.0',
				method: 'tools/call',
				params: {
					name: 'get_tasks',
					arguments: {
						filter: '"; DROP TABLE users; --',
					},
				},
				id: 1,
			};

			// Validation should pass (schema allows strings)
			// Injection prevention is handled by Vikunja backend, not MCP server
			const result = validateJsonRpcRequest(potentialInjection);
			expect(result.success).toBe(true);
		});

		it('should validate request with nested objects', () => {
			const nestedRequest = {
				jsonrpc: '2.0',
				method: 'tools/call',
				params: {
					deeply: {
						nested: {
							object: {
								structure: 'value',
							},
						},
					},
				},
				id: 1,
			};

			const result = validateJsonRpcRequest(nestedRequest);
			expect(result.success).toBe(true);
		});

		it('should validate request with arrays', () => {
			const arrayRequest = {
				jsonrpc: '2.0',
				method: 'tools/call',
				params: {
					items: [1, 2, 3, 4, 5],
				},
				id: 1,
			};

			const result = validateJsonRpcRequest(arrayRequest);
			expect(result.success).toBe(true);
		});
	});

	describe('Error Response Format', () => {
		it('should return structured error with validation details', () => {
			const invalidRequest = {
				jsonrpc: '1.0', // Wrong version
				method: 'test',
			};

			const result = validateJsonRpcRequest(invalidRequest);
			expect(result.success).toBe(false);
			if (!result.success) {
				expect(result.error).toHaveProperty('error');
				expect(result.error.error).toHaveProperty('code');
				expect(result.error.error).toHaveProperty('message');
				expect(result.error.error.data).toHaveProperty('validationErrors');
				expect(Array.isArray(result.error.error.data?.validationErrors)).toBe(true);
			}
		});

		it('should include field path in validation errors', () => {
			const invalidRequest = {
				jsonrpc: '2.0',
				// Missing method field
			};

			const result = validateJsonRpcRequest(invalidRequest);
			expect(result.success).toBe(false);
			if (!result.success && result.error.error.data?.validationErrors) {
				const methodError = result.error.error.data.validationErrors.find(
					(err) => err.path.includes('method')
				);
				expect(methodError).toBeDefined();
			}
		});
	});
});
