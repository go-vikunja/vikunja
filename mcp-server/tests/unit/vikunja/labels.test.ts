import { describe, it, expect, beforeEach, vi } from 'vitest';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { LabelAPI } from '../../../src/vikunja/labels.js';
import { VikunjaLabel } from '../../../src/vikunja/types.js';

describe('Label API Methods', () => {
  let mockClient: VikunjaClient;
  let labelAPI: LabelAPI;

  beforeEach(() => {
    mockClient = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      setToken: vi.fn(),
    } as unknown as VikunjaClient;

    labelAPI = new LabelAPI(mockClient);
  });

  describe('getLabels', () => {
    it('should get all labels', async () => {
      const mockLabels: VikunjaLabel[] = [
        {
          id: 1,
          title: 'Bug',
          description: 'Bug fix label',
          hex_color: '#ff0000',
          created: '2024-01-01T00:00:00Z',
          updated: '2024-01-01T00:00:00Z',
        },
        {
          id: 2,
          title: 'Feature',
          description: 'New feature label',
          hex_color: '#00ff00',
          created: '2024-01-02T00:00:00Z',
          updated: '2024-01-02T00:00:00Z',
        },
      ];

      vi.mocked(mockClient.get).mockResolvedValue(mockLabels);

      const labels = await labelAPI.getLabels();
      expect(labels).toEqual(mockLabels);
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/labels');
    });
  });

  describe('createLabel', () => {
    it('should create label', async () => {
      const input = {
        title: 'Enhancement',
        description: 'Improvement label',
        hex_color: '#0000ff',
      };

      const mockLabel: VikunjaLabel = {
        id: 3,
        title: input.title,
        description: input.description,
        hex_color: input.hex_color,
        created: '2024-01-03T00:00:00Z',
        updated: '2024-01-03T00:00:00Z',
      };

      vi.mocked(mockClient.put).mockResolvedValue(mockLabel);

      const label = await labelAPI.createLabel(input);
      expect(label).toEqual(mockLabel);
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/labels', input);
    });
  });

  describe('addLabelToTask', () => {
    it('should add label to task', async () => {
      vi.mocked(mockClient.put).mockResolvedValue({});

      await labelAPI.addLabelToTask(5, 1);
      expect(mockClient.put).toHaveBeenCalledWith('/api/v1/tasks/5/labels', { label_id: 1 });
    });
  });

  describe('removeLabelFromTask', () => {
    it('should remove label from task', async () => {
      vi.mocked(mockClient.delete).mockResolvedValue({});

      await labelAPI.removeLabelFromTask(5, 1);
      expect(mockClient.delete).toHaveBeenCalledWith('/api/v1/tasks/5/labels/1');
    });
  });
});
