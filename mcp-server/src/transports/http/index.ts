/**
 * HTTP Transport module exports
 * 
 * This module provides HTTP Streamable and SSE transport implementations
 * for the MCP protocol.
 */

// SSE transport (deprecated)
export { SSETransport } from './sse-transport.js';
export type { SSETransportConfig } from './sse-transport.js';

// HTTP Streamable transport (recommended)
export { HTTPStreamableTransport } from './http-streamable.js';
export type { HTTPStreamableTransportConfig } from './http-streamable.js';

// Health check endpoint
export { HealthCheckHandler } from './health-check.js';
export type { HealthCheckConfig } from './health-check.js';

// Session management
export { SessionManager, getSessionManager } from './session-manager.js';
export type {
	Session,
	SessionState,
	TransportType,
	ClientInfo,
} from './session-manager.js';

// Re-export everything for convenience
export * from './sse-transport.js';
export * from './http-streamable.js';
export * from './health-check.js';
export * from './session-manager.js';

