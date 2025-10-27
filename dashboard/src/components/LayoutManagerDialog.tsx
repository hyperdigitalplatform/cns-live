import React, { useState, useEffect } from 'react';
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
import {
  Settings,
  Trash2,
  Edit2,
  Globe,
  User,
  Grid,
  Zap,
  Loader2,
  Search,
  AlertCircle,
} from 'lucide-react';
import { api } from '@/services/api';
import type { LayoutPreferenceSummary, UpdateLayoutRequest } from '@/types';
import { cn } from '@/utils/cn';

interface LayoutManagerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onLayoutUpdate?: () => void;
}

export function LayoutManagerDialog({
  open,
  onOpenChange,
  onLayoutUpdate,
}: LayoutManagerDialogProps) {
  const [layouts, setLayouts] = useState<LayoutPreferenceSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [editingLayout, setEditingLayout] = useState<LayoutPreferenceSummary | null>(null);
  const [deletingLayout, setDeletingLayout] = useState<LayoutPreferenceSummary | null>(null);

  useEffect(() => {
    if (open) {
      loadLayouts();
    }
  }, [open]);

  const loadLayouts = async () => {
    setLoading(true);
    try {
      const response = await api.getLayouts();
      setLayouts(response.layouts);
    } catch (err) {
      console.error('Failed to load layouts:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (layout: LayoutPreferenceSummary) => {
    setEditingLayout(layout);
  };

  const handleDelete = (layout: LayoutPreferenceSummary) => {
    setDeletingLayout(layout);
  };

  const handleClose = () => {
    setSearchQuery('');
    setEditingLayout(null);
    setDeletingLayout(null);
    onOpenChange(false);
  };

  const filteredLayouts = layouts.filter((layout) => {
    const query = searchQuery.toLowerCase();
    return (
      layout.name.toLowerCase().includes(query) ||
      layout.description?.toLowerCase().includes(query) ||
      layout.created_by.toLowerCase().includes(query)
    );
  });

  const groupedLayouts = filteredLayouts.reduce(
    (acc, layout) => {
      if (layout.layout_type === 'standard') {
        acc.standard.push(layout);
      } else {
        acc.hotspot.push(layout);
      }
      return acc;
    },
    { standard: [] as LayoutPreferenceSummary[], hotspot: [] as LayoutPreferenceSummary[] }
  );

  return (
    <>
      <Dialog open={open && !editingLayout && !deletingLayout} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-3xl max-h-[85vh]" showClose onClose={handleClose}>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Settings className="h-5 w-5" />
              Manage Layouts
            </DialogTitle>
            <DialogDescription>
              View, edit, and delete saved layout configurations
            </DialogDescription>
          </DialogHeader>

          <DialogBody className="px-0">
            {/* Search */}
            <div className="px-6 mb-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search layouts..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 h-10 rounded-md border border-gray-300 bg-white text-sm
                    placeholder:text-gray-400
                    focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* Layouts List */}
            <div className="px-6 overflow-y-auto max-h-96">
              {loading ? (
                <div className="flex items-center justify-center py-12">
                  <Loader2 className="h-8 w-8 text-blue-600 animate-spin" />
                </div>
              ) : filteredLayouts.length === 0 ? (
                <div className="text-center py-12">
                  <Settings className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                  <p className="text-sm text-gray-500">
                    {searchQuery ? 'No layouts match your search' : 'No saved layouts yet'}
                  </p>
                </div>
              ) : (
                <div className="space-y-6">
                  {/* Standard Layouts */}
                  {groupedLayouts.standard.length > 0 && (
                    <div>
                      <div className="flex items-center gap-2 mb-3">
                        <Grid className="h-4 w-4 text-gray-500" />
                        <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wider">
                          Standard Layouts
                        </h3>
                        <span className="text-xs text-gray-400">
                          ({groupedLayouts.standard.length})
                        </span>
                      </div>
                      <div className="space-y-2">
                        {groupedLayouts.standard.map((layout) => (
                          <LayoutCard
                            key={layout.id}
                            layout={layout}
                            onEdit={handleEdit}
                            onDelete={handleDelete}
                          />
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Hotspot Layouts */}
                  {groupedLayouts.hotspot.length > 0 && (
                    <div>
                      <div className="flex items-center gap-2 mb-3">
                        <Zap className="h-4 w-4 text-gray-500" />
                        <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wider">
                          Hotspot Layouts
                        </h3>
                        <span className="text-xs text-gray-400">
                          ({groupedLayouts.hotspot.length})
                        </span>
                      </div>
                      <div className="space-y-2">
                        {groupedLayouts.hotspot.map((layout) => (
                          <LayoutCard
                            key={layout.id}
                            layout={layout}
                            onEdit={handleEdit}
                            onDelete={handleDelete}
                          />
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          </DialogBody>

          <DialogFooter>
            <Button variant="secondary" onClick={handleClose}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      {editingLayout && (
        <EditLayoutDialog
          layout={editingLayout}
          onClose={() => setEditingLayout(null)}
          onSuccess={() => {
            setEditingLayout(null);
            loadLayouts();
            onLayoutUpdate?.();
          }}
        />
      )}

      {/* Delete Confirmation Dialog */}
      {deletingLayout && (
        <DeleteLayoutDialog
          layout={deletingLayout}
          onClose={() => setDeletingLayout(null)}
          onSuccess={() => {
            setDeletingLayout(null);
            loadLayouts();
            onLayoutUpdate?.();
          }}
        />
      )}
    </>
  );
}

interface LayoutCardProps {
  layout: LayoutPreferenceSummary;
  onEdit: (layout: LayoutPreferenceSummary) => void;
  onDelete: (layout: LayoutPreferenceSummary) => void;
}

function LayoutCard({ layout, onEdit, onDelete }: LayoutCardProps) {
  return (
    <div className="group border border-gray-200 rounded-lg p-4 hover:border-gray-300 hover:shadow-sm transition-all">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <h4 className="text-sm font-semibold text-gray-900 truncate">
              {layout.name}
            </h4>
            <span
              className={cn(
                'inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium',
                layout.scope === 'global'
                  ? 'bg-blue-100 text-blue-700'
                  : 'bg-gray-100 text-gray-700'
              )}
            >
              {layout.scope === 'global' ? (
                <Globe className="h-3 w-3" />
              ) : (
                <User className="h-3 w-3" />
              )}
              {layout.scope === 'global' ? 'Global' : 'Personal'}
            </span>
          </div>
          {layout.description && (
            <p className="text-sm text-gray-600 mb-2">{layout.description}</p>
          )}
          <div className="flex items-center gap-4 text-xs text-gray-500">
            <span className="flex items-center gap-1">
              <Grid className="h-3 w-3" />
              {layout.camera_count} cameras
            </span>
            <span>•</span>
            <span>By {layout.created_by}</span>
            <span>•</span>
            <span>Updated {new Date(layout.updated_at).toLocaleDateString()}</span>
          </div>
        </div>

        <div className="flex items-center gap-1 ml-4 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={() => onEdit(layout)}
            className="p-2 text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
            title="Edit layout"
          >
            <Edit2 className="h-4 w-4" />
          </button>
          <button
            onClick={() => onDelete(layout)}
            className="p-2 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
            title="Delete layout"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );
}

interface EditLayoutDialogProps {
  layout: LayoutPreferenceSummary;
  onClose: () => void;
  onSuccess: () => void;
}

function EditLayoutDialog({ layout, onClose, onSuccess }: EditLayoutDialogProps) {
  const [name, setName] = useState(layout.name);
  const [description, setDescription] = useState(layout.description || '');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    if (!name.trim()) {
      setError('Layout name is required');
      return;
    }

    setSaving(true);
    setError(null);

    try {
      // Fetch full layout details to get cameras
      const fullLayout = await api.getLayout(layout.id);

      const updateRequest: UpdateLayoutRequest = {
        name: name.trim(),
        description: description.trim() || undefined,
        cameras: fullLayout.cameras || [],
      };

      await api.updateLayout(layout.id, updateRequest);
      onSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update layout');
      setSaving(false);
    }
  };

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-md" showClose onClose={onClose}>
        <DialogHeader>
          <DialogTitle>Edit Layout</DialogTitle>
          <DialogDescription>
            Update the layout name and description
          </DialogDescription>
        </DialogHeader>

        <DialogBody className="space-y-4">
          <Input
            label="Layout Name"
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
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={saving}
              rows={3}
              className="flex w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm
                placeholder:text-gray-400
                focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent
                disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>

          {error && (
            <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded-md p-3">
              {error}
            </div>
          )}
        </DialogBody>

        <DialogFooter>
          <Button variant="secondary" onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Button variant="primary" onClick={handleSave} disabled={saving || !name.trim()}>
            {saving ? 'Saving...' : 'Save Changes'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface DeleteLayoutDialogProps {
  layout: LayoutPreferenceSummary;
  onClose: () => void;
  onSuccess: () => void;
}

function DeleteLayoutDialog({ layout, onClose, onSuccess }: DeleteLayoutDialogProps) {
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleDelete = async () => {
    setDeleting(true);
    setError(null);

    try {
      await api.deleteLayout(layout.id);
      onSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete layout');
      setDeleting(false);
    }
  };

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-md" showClose onClose={onClose}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-red-600">
            <AlertCircle className="h-5 w-5" />
            Delete Layout
          </DialogTitle>
          <DialogDescription>
            This action cannot be undone
          </DialogDescription>
        </DialogHeader>

        <DialogBody>
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <p className="text-sm text-gray-700">
              Are you sure you want to delete{' '}
              <span className="font-semibold text-gray-900">{layout.name}</span>?
            </p>
            <p className="text-sm text-gray-600 mt-2">
              This layout contains {layout.camera_count} camera
              {layout.camera_count !== 1 ? 's' : ''} and will be permanently removed.
            </p>
          </div>

          {error && (
            <div className="mt-4 text-sm text-red-600 bg-red-50 border border-red-200 rounded-md p-3">
              {error}
            </div>
          )}
        </DialogBody>

        <DialogFooter>
          <Button variant="secondary" onClick={onClose} disabled={deleting}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleDelete} disabled={deleting}>
            {deleting ? (
              <>
                <span className="inline-block animate-spin mr-2">⏳</span>
                Deleting...
              </>
            ) : (
              <>
                <Trash2 className="h-4 w-4 mr-2" />
                Delete Layout
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
