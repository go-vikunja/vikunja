import { describe, it, expect } from 'vitest';
import { TransportType } from '../../src/config/index.js';

describe('Config Validation', () => {
  describe('TransportType enum', () => {
    it('should accept valid transport type "stdio"', () => {
      // Act
      const result = TransportType.safeParse('stdio');

      // Assert
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data).toBe('stdio');
      }
    });

    it('should accept valid transport type "http"', () => {
      // Act
      const result = TransportType.safeParse('http');

      // Assert
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data).toBe('http');
      }
    });

    it('should reject invalid transport type', () => {
      // Act
      const result = TransportType.safeParse('invalid');

      // Assert
      expect(result.success).toBe(false);
    });

    it('should reject empty transport type', () => {
      // Act
      const result = TransportType.safeParse('');

      // Assert
      expect(result.success).toBe(false);
    });

    it('should reject null transport type', () => {
      // Act
      const result = TransportType.safeParse(null);

      // Assert
      expect(result.success).toBe(false);
    });

    it('should reject undefined transport type', () => {
      // Act
      const result = TransportType.safeParse(undefined);

      // Assert
      expect(result.success).toBe(false);
    });
  });
});
