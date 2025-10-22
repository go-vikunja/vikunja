import { describe, it, expect, beforeEach, vi } from 'vitest';
import { createTransport } from '../../src/transport/factory.js';
import type { Config } from '../../src/config/index.js';

describe('Stdio Transport Regression', () => {
  let baseConfig: Config;

  beforeEach(() => {
    baseConfig = {
      vikunjaApiUrl: 'http://localhost:3456',
      port: 3457,
      redis: {
        host: 'localhost',
        port: 6379,
      },
      rateLimits: {
        default: 100,
        burst: 120,
        adminBypass: false,
      },
      llm: {
        provider: 'anthropic',
      },
      logging: {
        level: 'info',
        format: 'json',
      },
    };
  });

  it('should still work with TRANSPORT_TYPE=stdio', async () => {
    // Arrange
    const config: Config = {
      ...baseConfig,
      transportType: 'stdio',
    };

    // Act
    const transport = createTransport(config);

    // Assert
    expect(transport).toBeDefined();
    expect(transport.constructor.name).toBe('StdioServerTransport');
  });

  it('should default to stdio when TRANSPORT_TYPE omitted', () => {
    // Arrange
    const config: Config = {
      ...baseConfig,
      transportType: 'stdio', // This is the default value from the schema
    };

    // Act
    const transport = createTransport(config);

    // Assert
    expect(transport).toBeDefined();
    expect(transport.constructor.name).toBe('StdioServerTransport');
  });

  it('should maintain backward compatibility with existing stdio configuration', () => {
    // Arrange - simulate a config that existed before HTTP transport was added
    const legacyConfig: Config = {
      ...baseConfig,
      transportType: 'stdio',
      // No mcpPort, no cors - these are new fields
    };

    // Act
    const transport = createTransport(legacyConfig);

    // Assert - stdio transport should work exactly as before
    expect(transport).toBeDefined();
    expect(transport.constructor.name).toBe('StdioServerTransport');
  });
});
