import React, { useEffect, useState } from 'react';
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
import { Search, RefreshCw, Download, CheckCircle, Circle, Camera } from 'lucide-react';
import { cn } from '@/utils/cn';
import { api } from '@/services/api';
import type { DiscoveredCamera, ImportCameraRequest, CameraSource } from '@/types';

interface MilestoneCameraDiscoveryProps {
  open: boolean;
  onClose: () => void;
  onImport: (cameras: DiscoveredCamera[]) => void;
}

export function MilestoneCameraDiscovery({
  open,
  onClose,
  onImport,
}: MilestoneCameraDiscoveryProps) {
  const [cameras, setCameras] = useState<DiscoveredCamera[]>([]);
  const [loading, setLoading] = useState(false);
  const [importing, setImporting] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCameras, setSelectedCameras] = useState<Set<string>>(new Set());
  const [filterPTZ, setFilterPTZ] = useState<boolean | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [importResult, setImportResult] = useState<{
    imported: number;
    failed: number;
    errors?: string[];
  } | null>(null);

  // Fetch cameras from Milestone with ONVIF enrichment
  const fetchCameras = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.discoverCameras();
      setCameras(response.cameras || []);
    } catch (err) {
      console.error('Failed to fetch cameras:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch cameras');
    } finally {
      setLoading(false);
    }
  };

  // Import selected cameras
  const handleImport = async () => {
    const camerasToImport: ImportCameraRequest[] = cameras
      .filter((c) => selectedCameras.has(c.milestoneId))
      .map((c) => ({
        milestoneId: c.milestoneId,
        name: c.name,
        source: 'OTHER' as CameraSource,
        status: c.status as 'ONLINE' | 'OFFLINE',
        ptzEnabled: c.ptzEnabled,
        device: c.device,
        onvifEndpoint: c.onvifEndpoint,
        onvifUsername: c.onvifUsername,
        streams: c.streams,
        ptzCapabilities: c.ptzCapabilities,
      }));

    if (camerasToImport.length === 0) {
      setError('Please select cameras to import');
      return;
    }

    setImporting(true);
    setError(null);
    try {
      const response = await api.importCameras({ cameras: camerasToImport });
      setImportResult(response);

      if (response.imported > 0) {
        // Notify parent with successfully imported cameras
        const importedCameras = cameras.filter((c) => selectedCameras.has(c.milestoneId));

        // Call onImport and wait for camera list refresh to complete
        await Promise.resolve(onImport(importedCameras));

        setSelectedCameras(new Set());
      }
    } catch (err) {
      console.error('Failed to import cameras:', err);
      setError(err instanceof Error ? err.message : 'Import failed');
    } finally {
      setImporting(false);
    }
  };

  // Toggle camera selection
  const toggleCamera = (cameraId: string) => {
    const newSelected = new Set(selectedCameras);
    if (newSelected.has(cameraId)) {
      newSelected.delete(cameraId);
    } else {
      newSelected.add(cameraId);
    }
    setSelectedCameras(newSelected);
  };

  // Select all filtered cameras
  const selectAllFiltered = () => {
    const filtered = getFilteredCameras();
    const newSelected = new Set(filtered.map(cam => cam.milestoneId));
    setSelectedCameras(newSelected);
  };

  // Deselect all
  const deselectAll = () => {
    setSelectedCameras(new Set());
  };

  // Filter cameras
  const getFilteredCameras = () => {
    return cameras.filter(camera => {
      // Search filter
      const matchesSearch = camera.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                           camera.milestoneId.toLowerCase().includes(searchQuery.toLowerCase()) ||
                           camera.device.ip.includes(searchQuery);

      // PTZ filter
      const matchesPTZ = filterPTZ === null ||
                        (camera.ptzCapabilities &&
                         (camera.ptzCapabilities.pan || camera.ptzCapabilities.tilt || camera.ptzCapabilities.zoom)) === filterPTZ;

      return matchesSearch && matchesPTZ;
    });
  };

  const filteredCameras = getFilteredCameras();
  const selectedCount = selectedCameras.size;

  useEffect(() => {
    if (open) {
      fetchCameras();
      setSelectedCameras(new Set());
      setError(null);
      setImportResult(null);
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent onClose={onClose} className="max-w-4xl">
        <DialogHeader>
          <DialogTitle>Discover Milestone Cameras</DialogTitle>
          <DialogDescription>
            Import cameras from Milestone XProtect with ONVIF enrichment
          </DialogDescription>
        </DialogHeader>

        <DialogBody>
          {/* Error Message */}
          {error && (
            <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg text-red-800 text-sm">
              <strong>Error:</strong> {error}
            </div>
          )}

          {/* Import Result */}
          {importResult && (
            <div
              className={`mb-4 p-4 rounded-lg text-sm ${
                importResult.failed === 0
                  ? 'bg-green-50 border border-green-200 text-green-800'
                  : 'bg-yellow-50 border border-yellow-200 text-yellow-800'
              }`}
            >
              <strong>Import Complete:</strong> {importResult.imported} imported,{' '}
              {importResult.failed} failed
              {importResult.errors && importResult.errors.length > 0 && (
                <ul className="mt-2 ml-4 list-disc">
                  {importResult.errors.map((err, idx) => (
                    <li key={idx}>{err}</li>
                  ))}
                </ul>
              )}
            </div>
          )}

          {/* Search and Filters */}
          <div className="space-y-4 mb-4">
            <div className="flex items-center gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search cameras by name, ID, or IP..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-9 pr-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <Button
                variant="secondary"
                onClick={fetchCameras}
                disabled={loading}
                className="flex items-center gap-2"
              >
                <RefreshCw className={cn("w-4 h-4", loading && "animate-spin")} />
                Refresh
              </Button>
            </div>

            {/* Filters */}
            <div className="flex items-center gap-4 text-sm">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={filterPTZ === true}
                  onChange={(e) => setFilterPTZ(e.target.checked ? true : null)}
                  className="rounded"
                />
                <span>PTZ Capable</span>
              </label>
            </div>
          </div>

          {/* Selection Actions */}
          {selectedCount > 0 && (
            <div className="flex items-center justify-between p-3 bg-blue-50 rounded-lg mb-4">
              <span className="text-sm text-blue-900 font-medium">
                {selectedCount} camera{selectedCount !== 1 ? 's' : ''} selected
              </span>
              <div className="flex gap-2">
                <Button variant="secondary" size="sm" onClick={deselectAll}>
                  Deselect All
                </Button>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleImport}
                  disabled={importing}
                >
                  {importing ? 'Importing...' : 'Import Selected'}
                </Button>
              </div>
            </div>
          )}

          {/* Camera List */}
          <div className="border border-gray-200 rounded-lg overflow-hidden">
            <div className="max-h-96 overflow-y-auto">
              {loading ? (
                <div className="flex items-center justify-center py-12">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
                </div>
              ) : filteredCameras.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-12 text-gray-500">
                  <Camera className="w-12 h-12 mb-3 text-gray-400" />
                  <p className="text-sm">No cameras found</p>
                  <p className="text-xs mt-1">Click Refresh to discover cameras</p>
                </div>
              ) : (
                <div className="divide-y divide-gray-200">
                  {filteredCameras.map((camera) => (
                    <div
                      key={camera.milestoneId}
                      className="flex items-center gap-3 p-3 hover:bg-gray-50 transition-colors"
                    >
                      <input
                        type="checkbox"
                        checked={selectedCameras.has(camera.milestoneId)}
                        onChange={() => toggleCamera(camera.milestoneId)}
                        className="rounded"
                      />

                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-medium text-gray-900 truncate">
                            {camera.name}
                          </p>
                          {camera.status === 'ONLINE' && (
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                              <Circle className="w-2 h-2 fill-current" />
                              Online
                            </span>
                          )}
                        </div>
                        <p className="text-xs text-gray-500">
                          {camera.device.ip} • {camera.device.manufacturer} {camera.device.model}
                        </p>
                        <p className="text-xs text-gray-400">
                          {camera.streams.length} stream{camera.streams.length !== 1 ? 's' : ''} •
                          {camera.streams[0]?.resolution}
                        </p>
                      </div>

                      <div className="flex items-center gap-2">
                        {camera.recordingEnabled && (
                          <span className="flex items-center gap-1 text-xs text-green-600">
                            <Circle className="w-2 h-2 fill-current" />
                            Rec Enabled
                          </span>
                        )}
                        {(camera.ptzCapabilities.pan || camera.ptzCapabilities.tilt || camera.ptzCapabilities.zoom) && (
                          <span className="text-xs text-blue-600 font-medium">PTZ</span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Summary */}
          <div className="flex items-center justify-between mt-4 text-sm text-gray-600">
            <div>
              Showing {filteredCameras.length} of {cameras.length} cameras
            </div>
          </div>
        </DialogBody>

        <DialogFooter>
          <Button variant="secondary" onClick={onClose}>
            Close
          </Button>
          <Button
            variant="secondary"
            onClick={selectAllFiltered}
            disabled={filteredCameras.length === 0}
          >
            Select All Filtered
          </Button>
          <Button
            variant="primary"
            onClick={handleImport}
            disabled={selectedCount === 0 || loading || importing}
          >
            {importing ? 'Importing...' : `Import ${selectedCount > 0 ? `(${selectedCount})` : ''}`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
