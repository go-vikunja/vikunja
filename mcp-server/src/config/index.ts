import { z } from 'zod';

/**
 * Transport type enumeration
 */
export const TransportType = z.enum(['stdio', 'http']);
export type TransportType = z.infer<typeof TransportType>;

/**
 * Configuration schema validation
 */
const ConfigSchema = z.object({
  vikunjaApiUrl: z.string().url(),
  port: z.number().int().positive().default(3457),
  
  /**
   * MCP transport type
   * - stdio: Standard input/output (default, for subprocess communication)
   * - http: HTTP with Server-Sent Events (for remote clients)
   */
  transportType: TransportType.default('stdio'),
  
  /**
   * MCP server port (required for HTTP transport)
   * Blue environment: 3010
   * Green environment: 3011
   */
  mcpPort: z.number().int().positive().optional(),
  
  /**
   * CORS configuration for HTTP transport
   */
  cors: z
    .object({
      enabled: z.boolean().default(false),
      allowedOrigins: z.array(z.string().url()).default([]),
    })
    .optional(),
  
  redis: z.object({
    host: z.string().default('localhost'),
    port: z.number().int().positive().default(6379),
    password: z.string().optional(),
  }),
  rateLimits: z.object({
    default: z.number().int().positive().default(100),
    burst: z.number().int().positive().default(120),
    adminBypass: z.boolean().default(false),
  }),
  llm: z.object({
    provider: z.enum(['openai', 'anthropic', 'ollama']),
    apiKey: z.string().optional(),
    endpoint: z.string().url().optional(),
  }),
  logging: z.object({
    level: z.enum(['error', 'warn', 'info', 'debug']).default('info'),
    format: z.enum(['json', 'simple']).default('json'),
  }),
});

export type Config = z.infer<typeof ConfigSchema>;

/**
 * Load configuration from environment variables
 */
function loadConfig(): Config {
  const rawConfig = {
    vikunjaApiUrl: process.env['VIKUNJA_API_URL'] ?? 'http://localhost:3456',
    port: process.env['MCP_PORT'] ? parseInt(process.env['MCP_PORT'], 10) : 3457,
    transportType: (process.env['TRANSPORT_TYPE'] ?? 'stdio') as 'stdio' | 'http',
    mcpPort: process.env['MCP_PORT'] ? parseInt(process.env['MCP_PORT'], 10) : undefined,
    cors: process.env['CORS_ENABLED']
      ? {
          enabled: process.env['CORS_ENABLED'] === 'true',
          allowedOrigins: process.env['CORS_ALLOWED_ORIGINS']
            ? process.env['CORS_ALLOWED_ORIGINS'].split(',').map((url) => url.trim())
            : [],
        }
      : undefined,
    redis: {
      host: process.env['REDIS_HOST'] ?? 'localhost',
      port: process.env['REDIS_PORT'] ? parseInt(process.env['REDIS_PORT'], 10) : 6379,
      password: process.env['REDIS_PASSWORD'],
    },
    rateLimits: {
      default: process.env['RATE_LIMIT_DEFAULT']
        ? parseInt(process.env['RATE_LIMIT_DEFAULT'], 10)
        : 100,
      burst: process.env['RATE_LIMIT_BURST']
        ? parseInt(process.env['RATE_LIMIT_BURST'], 10)
        : 120,
      adminBypass: process.env['RATE_LIMIT_ADMIN_BYPASS'] === 'true',
    },
    llm: {
      provider: (process.env['LLM_PROVIDER'] ?? 'anthropic') as 'openai' | 'anthropic' | 'ollama',
      apiKey: process.env['LLM_API_KEY'],
      endpoint: process.env['LLM_ENDPOINT'],
    },
    logging: {
      level: (process.env['LOG_LEVEL'] ?? 'info') as 'error' | 'warn' | 'info' | 'debug',
      format: (process.env['LOG_FORMAT'] ?? 'json') as 'json' | 'simple',
    },
  };

  const parsed = ConfigSchema.parse(rawConfig);
  
  // Validate transport-specific configuration
  validateTransportConfig(parsed);
  
  return parsed;
}

/**
 * Validate configuration with cross-field constraints
 */
export function validateTransportConfig(config: Config): void {
  // Require mcpPort when using HTTP transport
  if (config.transportType === 'http' && (typeof config.mcpPort !== 'number' || config.mcpPort === 0)) {
    throw new Error(
      'Configuration error: MCP_PORT is required when TRANSPORT_TYPE=http'
    );
  }

  // Warn if CORS enabled without allowed origins
  if (config.cors?.enabled === true && config.cors.allowedOrigins.length === 0) {
    console.warn(
      'Warning: CORS enabled but no allowed origins configured. All origins will be denied.'
    );
  }
}

/**
 * Singleton configuration instance
 */
export const config = loadConfig();
