import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import type { Transport } from '@modelcontextprotocol/sdk/shared/transport.js';
import type { Config } from '../config/index.js';

/**
 * Create appropriate transport based on configuration
 * 
 * Note: For HTTP transport, this factory only validates configuration.
 * Actual HTTP server initialization happens in server.ts via startHttpTransport()
 */
export function createTransport(config: Config): Transport {
  if (config.transportType === 'stdio') {
    return new StdioServerTransport();
  } else if (config.transportType === 'http') {
    // HTTP transport requires separate Express server initialization
    // This is handled in server.ts via startHttpTransport() method
    throw new Error(
      'HTTP transport requires server.startHttpTransport() - use createTransport() only for stdio'
    );
  } else {
    const exhaustiveCheck: never = config.transportType;
    throw new Error(
      `Unsupported transport type: ${String(exhaustiveCheck)}. Expected 'stdio' or 'http'.`
    );
  }
}

/**
 * Validate transport configuration
 */
export function validateTransportConfig(config: Config): void {
  if (config.transportType === 'http' && (typeof config.mcpPort !== 'number' || config.mcpPort === 0)) {
    throw new Error('Configuration error: MCP_PORT is required when TRANSPORT_TYPE=http');
  }
}
