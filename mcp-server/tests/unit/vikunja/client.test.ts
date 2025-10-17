import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import axios, { AxiosError } from 'axios';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { AuthenticationError, NotFoundError, InternalError } from '../../../src/utils/errors.js';

// Mock axios
vi.mock('axios');

// Mock config
vi.mock('../../../src/config/index.js', () => ({
  config: {
    vikunjaApiUrl: 'http://localhost:3456',
    logging: {
      level: 'error',
      format: 'json',
    },
  },
}));

// Mock logger to suppress logs during tests
vi.mock('../../../src/utils/logger.js', () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

describe('VikunjaClient', () => {
  let client: VikunjaClient;
  let mockAxiosInstance: any;

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Create mock axios instance
    mockAxiosInstance = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    };
    
    // Mock axios.create to return our mock instance
    vi.mocked(axios.create).mockReturnValue(mockAxiosInstance as any);
    
    client = new VikunjaClient();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('setToken', () => {
    it('should set authentication token', () => {
      client.setToken('test-token-123');
      // Token is set internally, we'll verify it in subsequent requests
      expect(client).toBeDefined();
    });
  });

  describe('GET requests', () => {
    it('should make GET request with authentication token', async () => {
      const mockData = { id: 1, name: 'Test Project' };
      mockAxiosInstance.get.mockResolvedValue({ data: mockData });
      
      client.setToken('test-token');
      const result = await client.get('/api/v1/projects/1');
      
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        '/api/v1/projects/1',
        expect.objectContaining({
          headers: {
            Authorization: 'Bearer test-token',
          },
        })
      );
      expect(result).toEqual(mockData);
    });

    it('should make GET request with query params', async () => {
      const mockData = [{ id: 1 }, { id: 2 }];
      mockAxiosInstance.get.mockResolvedValue({ data: mockData });
      
      const result = await client.get('/api/v1/projects', { page: 1, limit: 10 });
      
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        '/api/v1/projects',
        expect.objectContaining({
          params: { page: 1, limit: 10 },
        })
      );
      expect(result).toEqual(mockData);
    });

    it('should timeout after 5 seconds', async () => {
      // Verify axios instance was created with 5000ms timeout
      expect(axios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          timeout: 5000,
        })
      );
    });
  });

  describe('POST requests', () => {
    it('should make POST request with data', async () => {
      const postData = { title: 'New Project' };
      const mockResponse = { id: 5, title: 'New Project' };
      mockAxiosInstance.post.mockResolvedValue({ data: mockResponse });
      
      client.setToken('test-token');
      const result = await client.post('/api/v1/projects', postData);
      
      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        '/api/v1/projects',
        postData,
        expect.objectContaining({
          headers: {
            Authorization: 'Bearer test-token',
          },
        })
      );
      expect(result).toEqual(mockResponse);
    });
  });

  describe('Error handling', () => {
    it('should not retry on 4xx errors', async () => {
      const error = Object.assign(new Error('Request failed with status code 401'), {
        isAxiosError: true,
        response: {
          status: 401,
          data: { message: 'Unauthorized' },
        },
      });
      
      mockAxiosInstance.get.mockRejectedValue(error);
      
      await expect(client.get('/api/v1/projects')).rejects.toThrow(AuthenticationError);
      
      // Should only be called once (no retries)
      expect(mockAxiosInstance.get).toHaveBeenCalledTimes(1);
    });

    it('should map 404 to NotFoundError', async () => {
      const error = Object.assign(new Error('Request failed with status code 404'), {
        isAxiosError: true,
        response: {
          status: 404,
          data: { message: 'Project not found' },
        },
      });
      
      mockAxiosInstance.get.mockRejectedValue(error);
      
      await expect(client.get('/api/v1/projects/999')).rejects.toThrow(NotFoundError);
    });

    it('should retry on network errors', async () => {
      const networkError = Object.assign(new Error('Network Error'), {
        isAxiosError: true,
        code: 'ECONNREFUSED',
      });
      
      const successResponse = { data: { id: 1 } };
      
      // Fail twice, then succeed
      mockAxiosInstance.get
        .mockRejectedValueOnce(networkError)
        .mockRejectedValueOnce(networkError)
        .mockResolvedValueOnce(successResponse);
      
      const result = await client.get('/api/v1/projects/1');
      
      expect(result).toEqual({ id: 1 });
      // Should be called 3 times (initial + 2 retries)
      expect(mockAxiosInstance.get).toHaveBeenCalledTimes(3);
    });

    it('should retry on 5xx errors', async () => {
      const serverError = Object.assign(new Error('Request failed with status code 500'), {
        isAxiosError: true,
        response: {
          status: 500,
          data: { message: 'Internal Server Error' },
        },
      });
      
      const successResponse = { data: { id: 1 } };
      
      // Fail once, then succeed
      mockAxiosInstance.get
        .mockRejectedValueOnce(serverError)
        .mockResolvedValueOnce(successResponse);
      
      const result = await client.get('/api/v1/projects/1');
      
      expect(result).toEqual({ id: 1 });
      expect(mockAxiosInstance.get).toHaveBeenCalledTimes(2);
    });

    it('should fail after max retries', async () => {
      const networkError = Object.assign(new Error('Network Error'), {
        isAxiosError: true,
        code: 'ECONNREFUSED',
      });
      
      // Always fail
      mockAxiosInstance.get.mockRejectedValue(networkError);
      
      await expect(client.get('/api/v1/projects/1')).rejects.toThrow(InternalError);
      
      // Should be called 4 times (initial + 3 retries)
      expect(mockAxiosInstance.get).toHaveBeenCalledTimes(4);
    }, 10000); // Increase timeout for retries
  });

  describe('PUT requests', () => {
    it('should make PUT request with data', async () => {
      const updateData = { title: 'Updated Project' };
      const mockResponse = { id: 1, title: 'Updated Project' };
      mockAxiosInstance.put.mockResolvedValue({ data: mockResponse });
      
      client.setToken('test-token');
      const result = await client.put('/api/v1/projects/1', updateData);
      
      expect(mockAxiosInstance.put).toHaveBeenCalledWith(
        '/api/v1/projects/1',
        updateData,
        expect.objectContaining({
          headers: {
            Authorization: 'Bearer test-token',
          },
        })
      );
      expect(result).toEqual(mockResponse);
    });
  });

  describe('DELETE requests', () => {
    it('should make DELETE request', async () => {
      mockAxiosInstance.delete.mockResolvedValue({ data: {} });
      
      client.setToken('test-token');
      await client.delete('/api/v1/projects/1');
      
      expect(mockAxiosInstance.delete).toHaveBeenCalledWith(
        '/api/v1/projects/1',
        expect.objectContaining({
          headers: {
            Authorization: 'Bearer test-token',
          },
        })
      );
    });
  });
});
