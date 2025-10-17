/**
 * Base MCP error class with JSON-RPC code
 */
export class MCPError extends Error {
  constructor(
    message: string,
    public code: number,
    public data?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'MCPError';
  }
}

/**
 * Authentication error (code: -32000)
 */
export class AuthenticationError extends MCPError {
  constructor(message = 'Authentication failed', data?: Record<string, unknown>) {
    super(message, -32000, data);
    this.name = 'AuthenticationError';
  }
}

/**
 * Permission error (code: -32000)
 */
export class PermissionError extends MCPError {
  constructor(message = 'Permission denied', data?: Record<string, unknown>) {
    super(message, -32000, data);
    this.name = 'PermissionError';
  }
}

/**
 * Not found error (code: -32000)
 */
export class NotFoundError extends MCPError {
  constructor(message = 'Resource not found', data?: Record<string, unknown>) {
    super(message, -32000, data);
    this.name = 'NotFoundError';
  }
}

/**
 * Validation error (code: -32602)
 */
export class ValidationError extends MCPError {
  constructor(message = 'Invalid parameters', data?: Record<string, unknown>) {
    super(message, -32602, data);
    this.name = 'ValidationError';
  }
}

/**
 * Rate limit error (code: -32000)
 */
export class RateLimitError extends MCPError {
  constructor(
    message = 'Rate limit exceeded',
    public remaining: number,
    public resetAt: Date,
    data?: Record<string, unknown>
  ) {
    super(message, -32000, { ...data, remaining, resetAt: resetAt.toISOString() });
    this.name = 'RateLimitError';
  }
}

/**
 * Internal error (code: -32603)
 */
export class InternalError extends MCPError {
  constructor(message = 'Internal server error', data?: Record<string, unknown>) {
    super(message, -32603, data);
    this.name = 'InternalError';
  }
}

/**
 * Vikunja API error structure
 */
export interface VikunjaApiError {
  response:
    | {
        status: number;
        data?: {
          message?: string;
          code?: number;
        };
      }
    | undefined;
  message: string;
}

/**
 * Map Vikunja API error to MCP error
 */
export function mapVikunjaError(apiError: VikunjaApiError): MCPError {
  const status = apiError.response?.status;
  const message = apiError.response?.data?.message ?? apiError.message;

  switch (status) {
    case 401:
      return new AuthenticationError(message, { vikunjaError: message });
    case 403:
      return new PermissionError(message, { vikunjaError: message });
    case 404:
      return new NotFoundError(message, { vikunjaError: message });
    case 422:
      return new ValidationError(message, { vikunjaError: message });
    case 429:
      return new RateLimitError(message, 0, new Date(), { vikunjaError: message });
    default:
      return new InternalError(message, { vikunjaError: message, status });
  }
}

/**
 * JSON-RPC error format
 */
export interface JSONRPCError {
  code: number;
  message: string;
  data?: Record<string, unknown>;
}

/**
 * Format error for JSON-RPC response
 */
export function formatErrorForMCP(error: Error): JSONRPCError {
  if (error instanceof MCPError) {
    const result: JSONRPCError = {
      code: error.code,
      message: error.message,
    };
    if (error.data) {
      result.data = error.data;
    }
    return result;
  }

  // Unknown error - treat as internal error
  return {
    code: -32603,
    message: error.message || 'Internal server error',
    data: {
      error: error.name,
    },
  };
}
