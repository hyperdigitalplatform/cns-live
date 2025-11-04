import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogBody,
  DialogFooter,
  Button,
  Input,
} from './ui/Dialog';
import { Save, Globe, User } from 'lucide-react';
import { api } from '@/services/api';
import type { LayoutType, LayoutScope, LayoutCameraAssignment } from '@/types';

interface SaveLayoutDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  layoutType: LayoutType;
  gridLayout: string; // "2x2", "3x3", "9-way-1-hotspot", etc.
  cameras: Array<{ camera_id: string; position_index: number }>;
  onSuccess?: () => void;
}

export function SaveLayoutDialog({
  open,
  onOpenChange,
  layoutType,
  gridLayout,
  cameras,
  onSuccess,
}: SaveLayoutDialogProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [scope, setScope] = useState<LayoutScope>('local');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    if (!name.trim()) {
      setError('Layout name is required');
      return;
    }

    if (cameras.length === 0) {
      setError('At least one camera must be added to the layout');
      return;
    }

    setSaving(true);
    setError(null);

    try {
      const cameraAssignments: LayoutCameraAssignment[] = cameras.map((cam) => ({
        camera_id: cam.camera_id,
        position_index: cam.position_index,
      }));

      console.log('Creating layout with grid_layout:', gridLayout, 'layoutType:', layoutType);

      await api.createLayout({
        name: name.trim(),
        description: description.trim() || undefined,
        layout_type: layoutType,
        grid_layout: gridLayout,
        scope,
        created_by: 'current-user', // TODO: Get from auth context
        cameras: cameraAssignments,
      });

      // Reset form
      setName('');
      setDescription('');
      setScope('local');
      setError(null);

      onOpenChange(false);
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save layout');
    } finally {
      setSaving(false);
    }
  };

  const handleClose = () => {
    if (!saving) {
      setName('');
      setDescription('');
      setScope('local');
      setError(null);
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md" showClose onClose={handleClose}>
        <DialogHeader>
          <DialogTitle>Save Layout</DialogTitle>
          <DialogDescription>
            Save the current {layoutType} layout configuration for future use
          </DialogDescription>
        </DialogHeader>

        <DialogBody className="space-y-4">
          <Input
            label="Layout Name"
            placeholder="e.g., Main Control Room"
            value={name}
            onChange={(e) => setName(e.target.value)}
            error={error && !name.trim() ? error : undefined}
            disabled={saving}
            autoFocus
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">
              Description (Optional)
            </label>
            <textarea
              placeholder="Add a description for this layout..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={saving}
              rows={3}
              className="flex w-full rounded-md border border-gray-300 dark:border-dark-border bg-white dark:bg-dark-surface px-3 py-2 text-sm text-gray-900 dark:text-text-primary
                placeholder:text-gray-400 dark:placeholder:text-text-muted
                focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent
                disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Visibility
            </label>
            <div className="space-y-2">
              <label
                className={`flex items-center p-3 border rounded-md cursor-pointer transition-colors ${
                  scope === 'local'
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30'
                    : 'border-gray-300 dark:border-dark-border hover:bg-gray-50 dark:hover:bg-dark-surface'
                }`}
              >
                <input
                  type="radio"
                  name="scope"
                  value="local"
                  checked={scope === 'local'}
                  onChange={(e) => setScope(e.target.value as LayoutScope)}
                  disabled={saving}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                />
                <User className="ml-3 h-4 w-4 text-gray-600" />
                <div className="ml-2">
                  <div className="text-sm font-medium text-gray-900">
                    Personal
                  </div>
                  <div className="text-xs text-gray-500">
                    Only visible to you
                  </div>
                </div>
              </label>

              <label
                className={`flex items-center p-3 border rounded-md cursor-pointer transition-colors ${
                  scope === 'global'
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30'
                    : 'border-gray-300 dark:border-dark-border hover:bg-gray-50 dark:hover:bg-dark-surface'
                }`}
              >
                <input
                  type="radio"
                  name="scope"
                  value="global"
                  checked={scope === 'global'}
                  onChange={(e) => setScope(e.target.value as LayoutScope)}
                  disabled={saving}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                />
                <Globe className="ml-3 h-4 w-4 text-gray-600" />
                <div className="ml-2">
                  <div className="text-sm font-medium text-gray-900">
                    Global
                  </div>
                  <div className="text-xs text-gray-500">
                    Visible to all users
                  </div>
                </div>
              </label>
            </div>
          </div>

          <div className="bg-gray-50 dark:bg-dark-surface border border-gray-200 dark:border-dark-border rounded-md p-3">
            <div className="text-xs text-gray-600 dark:text-text-secondary">
              <div className="flex items-center justify-between">
                <span>Layout Type:</span>
                <span className="font-medium text-gray-900 dark:text-text-primary capitalize">
                  {layoutType}
                </span>
              </div>
              <div className="flex items-center justify-between mt-1">
                <span>Cameras:</span>
                <span className="font-medium text-gray-900 dark:text-text-primary">
                  {cameras.length}
                </span>
              </div>
            </div>
          </div>

          {error && cameras.length === 0 && (
            <div className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 rounded-md p-3">
              {error}
            </div>
          )}
        </DialogBody>

        <DialogFooter>
          <Button
            variant="secondary"
            onClick={handleClose}
            disabled={saving}
          >
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={handleSave}
            disabled={saving || !name.trim()}
          >
            {saving ? (
              <>
                <span className="inline-block animate-spin mr-2">⏳</span>
                Saving...
              </>
            ) : (
              <>
                <Save className="h-4 w-4 mr-2" />
                Save Layout
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
