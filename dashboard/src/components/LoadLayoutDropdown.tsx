import React, { useState, useEffect, useRef } from 'react';
import { FolderOpen, ChevronDown, Globe, User, Grid, Zap, Loader2 } from 'lucide-react';
import { api } from '@/services/api';
import type { LayoutPreferenceSummary, LayoutType, LayoutScope } from '@/types';
import { cn } from '@/utils/cn';

interface LoadLayoutDropdownProps {
  onLayoutSelect: (layout: LayoutPreferenceSummary) => void;
  currentLayoutType?: LayoutType;
}

export function LoadLayoutDropdown({
  onLayoutSelect,
  currentLayoutType,
}: LoadLayoutDropdownProps) {
  const [open, setOpen] = useState(false);
  const [layouts, setLayouts] = useState<LayoutPreferenceSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setOpen(false);
      }
    };

    if (open) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [open]);

  useEffect(() => {
    if (open && layouts.length === 0) {
      loadLayouts();
    }
  }, [open]);

  const loadLayouts = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await api.getLayouts();
      setLayouts(response.layouts);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load layouts');
    } finally {
      setLoading(false);
    }
  };

  const handleSelect = (layout: LayoutPreferenceSummary) => {
    onLayoutSelect(layout);
    setOpen(false);
  };

  const groupedLayouts = layouts.reduce(
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
    <div ref={dropdownRef} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className={cn(
          'inline-flex items-center justify-center gap-2',
          'px-4 h-10 rounded-md border border-gray-300 dark:border-dark-border bg-white dark:bg-dark-surface',
          'text-sm font-medium text-gray-700 dark:text-text-secondary',
          'hover:bg-gray-50 dark:hover:bg-dark-elevated hover:border-gray-400 dark:hover:border-dark-border',
          'focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-600 focus:ring-offset-2',
          'transition-colors',
          open && 'bg-gray-50 dark:bg-dark-elevated border-gray-400 dark:border-dark-border'
        )}
      >
        <FolderOpen className="h-4 w-4" />
        Load Layout
        <ChevronDown
          className={cn(
            'h-4 w-4 transition-transform',
            open && 'transform rotate-180'
          )}
        />
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-80 bg-white dark:bg-dark-sidebar rounded-lg shadow-lg border border-gray-200 dark:border-dark-border z-50 animate-in fade-in slide-in-from-top-2 duration-200">
          <div className="p-3 border-b border-gray-200 dark:border-dark-border">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-text-primary">
              Saved Layouts
            </h3>
            <p className="text-xs text-gray-500 dark:text-text-muted mt-0.5">
              Select a layout to apply
            </p>
          </div>

          <div className="max-h-96 overflow-y-auto">
            {loading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 text-blue-600 animate-spin" />
                <span className="ml-2 text-sm text-gray-600">
                  Loading layouts...
                </span>
              </div>
            ) : error ? (
              <div className="p-4 text-sm text-red-600 text-center">
                {error}
                <button
                  onClick={loadLayouts}
                  className="block w-full mt-2 text-blue-600 hover:text-blue-700"
                >
                  Try again
                </button>
              </div>
            ) : layouts.length === 0 ? (
              <div className="p-8 text-center">
                <FolderOpen className="h-12 w-12 text-gray-300 dark:text-text-muted mx-auto mb-3" />
                <p className="text-sm text-gray-500 dark:text-text-muted">No saved layouts yet</p>
                <p className="text-xs text-gray-400 dark:text-text-muted mt-1">
                  Save your first layout to get started
                </p>
              </div>
            ) : (
              <div className="py-2">
                {/* Standard Layouts */}
                {groupedLayouts.standard.length > 0 && (
                  <div className="mb-2">
                    <div className="px-3 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider flex items-center gap-2">
                      <Grid className="h-3 w-3" />
                      Standard Layouts
                    </div>
                    {groupedLayouts.standard.map((layout) => (
                      <LayoutItem
                        key={layout.id}
                        layout={layout}
                        onSelect={handleSelect}
                        isCurrentType={currentLayoutType === 'standard'}
                      />
                    ))}
                  </div>
                )}

                {/* Hotspot Layouts */}
                {groupedLayouts.hotspot.length > 0 && (
                  <div>
                    <div className="px-3 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider flex items-center gap-2">
                      <Zap className="h-3 w-3" />
                      Hotspot Layouts
                    </div>
                    {groupedLayouts.hotspot.map((layout) => (
                      <LayoutItem
                        key={layout.id}
                        layout={layout}
                        onSelect={handleSelect}
                        isCurrentType={currentLayoutType === 'hotspot'}
                      />
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

interface LayoutItemProps {
  layout: LayoutPreferenceSummary;
  onSelect: (layout: LayoutPreferenceSummary) => void;
  isCurrentType: boolean;
}

function LayoutItem({ layout, onSelect, isCurrentType }: LayoutItemProps) {
  return (
    <button
      onClick={() => onSelect(layout)}
      className={cn(
        'w-full px-3 py-2.5 text-left hover:bg-gray-50 dark:hover:bg-dark-surface transition-colors',
        'flex items-start gap-3 group'
      )}
    >
      <div
        className={cn(
          'mt-0.5 p-1.5 rounded',
          layout.scope === 'global'
            ? 'bg-blue-100 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400'
            : 'bg-gray-100 dark:bg-dark-surface text-gray-600 dark:text-text-secondary'
        )}
      >
        {layout.scope === 'global' ? (
          <Globe className="h-3.5 w-3.5" />
        ) : (
          <User className="h-3.5 w-3.5" />
        )}
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium text-gray-900 dark:text-text-primary truncate group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
            {layout.name}
          </p>
          {isCurrentType && (
            <span className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-400">
              Current Type
            </span>
          )}
        </div>
        {layout.description && (
          <p className="text-xs text-gray-500 dark:text-text-muted mt-0.5 truncate">
            {layout.description}
          </p>
        )}
        <div className="flex items-center gap-3 mt-1.5 text-xs text-gray-400 dark:text-text-muted">
          <span className="flex items-center gap-1">
            <Grid className="h-3 w-3" />
            {layout.camera_count} cameras
          </span>
          <span className="text-gray-300">•</span>
          <span>
            {new Date(layout.updated_at).toLocaleDateString()}
          </span>
        </div>
      </div>
    </button>
  );
}
