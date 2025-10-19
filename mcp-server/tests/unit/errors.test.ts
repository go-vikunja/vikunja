import { describe, it, expect } from 'vitest';
import {
  MCPError,
  AuthenticationError,
  PermissionError,
  NotFoundError,
  ValidationError,
  RateLimitError,
  InternalError,
  mapVikunjaError,
  formatErrorForMCP,
} from '../../src/utils/errors.js';

describe('Error Utilities', () => {
  describe('MCPError', () => {
    it('should create error with code and data', () => {
      const error = new MCPError('Test error', -32000, { detail: 'test' });
      expect(error.message).toBe('Test error');
      expect(error.code).toBe(-32000);
      expect(error.data).toEqual({ detail: 'test' });
    });
  });

  describe('mapVikunjaError', () => {
    it('should map Vikunja 401 to AuthenticationError', () => {
      const apiError = {
        response: {
          status: 401,
          data: { message: 'Invalid token' },
        },
        message: 'Request failed',
      };

      const error = mapVikunjaError(apiError);
      expect(error).toBeInstanceOf(AuthenticationError);
      expect(error.message).toBe('Invalid token');
      expect(error.data).toEqual({ vikunjaError: 'Invalid token' });
    });

    it('should map Vikunja 403 to PermissionError', () => {
      const apiError = {
        response: {
          status: 403,
          data: { message: 'Access denied' },
        },
        message: 'Request failed',
      };

      const error = mapVikunjaError(apiError);
      expect(error).toBeInstanceOf(PermissionError);
      expect(error.message).toBe('Access denied');
    });

    it('should map Vikunja 404 to NotFoundError', () => {
      const apiError = {
        response: {
          status: 404,
          data: { message: 'Project not found' },
        },
        message: 'Request failed',
      };

      const error = mapVikunjaError(apiError);
      expect(error).toBeInstanceOf(NotFoundError);
      expect(error.message).toBe('Project not found');
    });

    it('should map Vikunja 422 to ValidationError', () => {
      const apiError = {
        response: {
          status: 422,
          data: { message: 'Invalid input' },
        },
        message: 'Request failed',
      };

      const error = mapVikunjaError(apiError);
      expect(error).toBeInstanceOf(ValidationError);
      expect(error.message).toBe('Invalid input');
    });

    it('should map unknown status to InternalError', () => {
      const apiError = {
        response: {
          status: 500,
          data: { message: 'Server error' },
        },
        message: 'Request failed',
      };

      const error = mapVikunjaError(apiError);
      expect(error).toBeInstanceOf(InternalError);
    });
  });

  describe('formatErrorForMCP', () => {
    it('should format MCP error for JSON-RPC', () => {
      const error = new PermissionError('Access denied', { projectId: 5 });
      const formatted = formatErrorForMCP(error);

      expect(formatted).toEqual({
        code: -32000,
        message: 'Access denied',
        data: { projectId: 5 },
      });
    });

    it('should format unknown error with default code', () => {
      const error = new Error('Something went wrong');
      const formatted = formatErrorForMCP(error);

      expect(formatted).toEqual({
        code: -32603,
        message: 'Something went wrong',
        data: { error: 'Error' },
      });
    });

    it('should include data payload in error', () => {
      const error = new NotFoundError('Task not found', { taskId: 123 });
      const formatted = formatErrorForMCP(error);

      expect(formatted.data).toEqual({ taskId: 123 });
    });
  });

  describe('RateLimitError', () => {
    it('should include remaining and resetAt', () => {
      const resetAt = new Date('2025-10-17T12:00:00Z');
      const error = new RateLimitError('Too many requests', 0, resetAt);

      expect(error.remaining).toBe(0);
      expect(error.resetAt).toEqual(resetAt);
      expect(error.data?.['resetAt']).toBe('2025-10-17T12:00:00.000Z');
    });
  });
});
