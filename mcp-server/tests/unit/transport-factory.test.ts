import { describe, it, expect } from 'vitest';
import { createTransport, validateTransportConfig } from '../../src/transport/factory.js';
import type { Config } from '../../src/config/index.js';

describe('Transport Factory', () => {
  describe('createTransport', () => {
    it('should return StdioServerTransport for stdio transport type', async () => {
      // Arrange
      const config: Config = {
        vikunjaApiUrl: 'http://localhost:3456',
        port: 3457,
        transportType: 'stdio',
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

      // Act
      const transport = createTransport(config);

      // Assert
      expect(transport).toBeDefined();
      expect(transport.constructor.name).toBe('StdioServerTransport');
    });

    it('should throw error for HTTP transport (requires explicit server initialization)', () => {
      // Arrange
      const config: Config = {
        vikunjaApiUrl: 'http://localhost:3456',
        port: 3457,
        transportType: 'http',
        mcpPort: 3010,
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

      // Act & Assert
      expect(() => createTransport(config)).toThrow(
        'HTTP transport requires server.startHttpTransport()'
      );
    });
  });

  describe('validateTransportConfig', () => {
    it('should require MCP_PORT for HTTP transport', () => {
      // Arrange
      const config: Config = {
        vikunjaApiUrl: 'http://localhost:3456',
        port: 3457,
        transportType: 'http',
        // mcpPort is missing
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

      // Act & Assert
      expect(() => validateTransportConfig(config)).toThrow(
        'MCP_PORT is required when TRANSPORT_TYPE=http'
      );
    });

    it('should not throw for HTTP transport with mcpPort', () => {
      // Arrange
      const config: Config = {
        vikunjaApiUrl: 'http://localhost:3456',
        port: 3457,
        transportType: 'http',
        mcpPort: 3010,
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

      // Act & Assert
      expect(() => validateTransportConfig(config)).not.toThrow();
    });

    it('should not throw for stdio transport without mcpPort', () => {
      // Arrange
      const config: Config = {
        vikunjaApiUrl: 'http://localhost:3456',
        port: 3457,
        transportType: 'stdio',
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

      // Act & Assert
      expect(() => validateTransportConfig(config)).not.toThrow();
    });
  });
});
