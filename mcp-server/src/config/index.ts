import { z } from 'zod';

/**
 * Configuration schema validation
 */
const ConfigSchema = z.object({
  vikunjaApiUrl: z.string().url(),
  port: z.number().int().positive().default(3457),
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

  return ConfigSchema.parse(rawConfig);
}

/**
 * Singleton configuration instance
 */
export const config = loadConfig();
