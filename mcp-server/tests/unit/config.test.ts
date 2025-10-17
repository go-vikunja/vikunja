import { describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('Configuration', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    // Reset environment before each test
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    // Restore original environment
    process.env = originalEnv;
  });

  it('should load default config', async () => {
    // Set minimal required env vars
    process.env['VIKUNJA_API_URL'] = 'http://localhost:3456';

    // Dynamic import to get fresh config
    const { config } = await import('../../src/config/index.js');

    expect(config.vikunjaApiUrl).toBe('http://localhost:3456');
    expect(config.port).toBe(3457);
    expect(config.redis.host).toBe('localhost');
    expect(config.redis.port).toBe(6379);
    expect(config.rateLimits.default).toBe(100);
    expect(config.rateLimits.burst).toBe(120);
    expect(config.logging.level).toBe('info');
  });

  it('should override with env vars', async () => {
    process.env['VIKUNJA_API_URL'] = 'http://vikunja.example.com';
    process.env['MCP_PORT'] = '4000';
    process.env['REDIS_HOST'] = 'redis.example.com';
    process.env['REDIS_PORT'] = '6380';
    process.env['RATE_LIMIT_DEFAULT'] = '200';
    process.env['LOG_LEVEL'] = 'debug';

    const { config } = await import('../../src/config/index.js');

    expect(config.vikunjaApiUrl).toBe('http://vikunja.example.com');
    expect(config.port).toBe(4000);
    expect(config.redis.host).toBe('redis.example.com');
    expect(config.redis.port).toBe(6380);
    expect(config.rateLimits.default).toBe(200);
    expect(config.logging.level).toBe('debug');
  });

  it('should validate required fields', async () => {
    // Remove required field
    delete process.env['VIKUNJA_API_URL'];

    await expect(async () => {
      await import('../../src/config/index.js');
    }).rejects.toThrow();
  });

  it('should reject invalid values', async () => {
    process.env['VIKUNJA_API_URL'] = 'not-a-url';

    await expect(async () => {
      await import('../../src/config/index.js');
    }).rejects.toThrow();
  });
});
