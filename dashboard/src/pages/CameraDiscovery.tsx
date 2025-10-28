import { useState } from 'react';
import { api } from '@/services/api';
import type {
  DiscoveredCamera,
  CameraSource,
  ImportCameraRequest,
} from '@/types';

export default function CameraDiscovery() {
  const [cameras, setCameras] = useState<DiscoveredCamera[]>([]);
  const [selectedCameras, setSelectedCameras] = useState<Set<string>>(
    new Set()
  );
  const [isDiscovering, setIsDiscovering] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [importResult, setImportResult] = useState<{
    imported: number;
    failed: number;
    errors?: string[];
  } | null>(null);

  const handleDiscover = async () => {
    setIsDiscovering(true);
    setError(null);
    setImportResult(null);
    try {
      const response = await api.discoverCameras();
      setCameras(response.cameras);
      setSelectedCameras(new Set());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Discovery failed');
    } finally {
      setIsDiscovering(false);
    }
  };

  const toggleCamera = (cameraId: string) => {
    const newSelected = new Set(selectedCameras);
    if (newSelected.has(cameraId)) {
      newSelected.delete(cameraId);
    } else {
      newSelected.add(cameraId);
    }
    setSelectedCameras(newSelected);
  };

  const toggleAll = () => {
    if (selectedCameras.size === cameras.length) {
      setSelectedCameras(new Set());
    } else {
      setSelectedCameras(new Set(cameras.map((c) => c.milestoneId)));
    }
  };

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

    setIsImporting(true);
    setError(null);
    try {
      const response = await api.importCameras({ cameras: camerasToImport });
      setImportResult(response);
      if (response.imported > 0) {
        // Clear selected cameras on success
        setSelectedCameras(new Set());
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Import failed');
    } finally {
      setIsImporting(false);
    }
  };

  return (
    <div className="h-full flex flex-col bg-gray-900 p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white mb-2">Camera Discovery</h1>
        <p className="text-gray-400">
          Discover cameras from Milestone XProtect with ONVIF enrichment
        </p>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={handleDiscover}
          disabled={isDiscovering}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed flex items-center gap-2"
        >
          {isDiscovering ? (
            <>
              <span className="animate-spin">⚙️</span>
              Discovering...
            </>
          ) : (
            <>🔍 Discover Cameras</>
          )}
        </button>

        {cameras.length > 0 && (
          <button
            onClick={handleImport}
            disabled={isImporting || selectedCameras.size === 0}
            className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {isImporting ? (
              <>
                <span className="animate-spin">⚙️</span>
                Importing...
              </>
            ) : (
              <>
                ⬇️ Import Selected ({selectedCameras.size})
              </>
            )}
          </button>
        )}
      </div>

      {/* Error Message */}
      {error && (
        <div className="mb-4 p-4 bg-red-900/50 border border-red-700 rounded-lg text-red-200">
          <strong>Error:</strong> {error}
        </div>
      )}

      {/* Import Result */}
      {importResult && (
        <div
          className={`mb-4 p-4 rounded-lg ${
            importResult.failed === 0
              ? 'bg-green-900/50 border border-green-700 text-green-200'
              : 'bg-yellow-900/50 border border-yellow-700 text-yellow-200'
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

      {/* Camera Table */}
      {cameras.length > 0 && (
        <div className="flex-1 overflow-auto bg-gray-800 rounded-lg">
          <table className="w-full text-left text-sm">
            <thead className="bg-gray-700 text-gray-300 sticky top-0">
              <tr>
                <th className="p-3">
                  <input
                    type="checkbox"
                    checked={
                      cameras.length > 0 &&
                      selectedCameras.size === cameras.length
                    }
                    onChange={toggleAll}
                    className="w-4 h-4 cursor-pointer"
                  />
                </th>
                <th className="p-3">Name</th>
                <th className="p-3">IP Address</th>
                <th className="p-3">Manufacturer</th>
                <th className="p-3">Model</th>
                <th className="p-3">Firmware</th>
                <th className="p-3">Streams</th>
                <th className="p-3">PTZ</th>
                <th className="p-3">Status</th>
              </tr>
            </thead>
            <tbody className="text-gray-300">
              {cameras.map((camera) => (
                <tr
                  key={camera.milestoneId}
                  className="border-b border-gray-700 hover:bg-gray-750"
                >
                  <td className="p-3">
                    <input
                      type="checkbox"
                      checked={selectedCameras.has(camera.milestoneId)}
                      onChange={() => toggleCamera(camera.milestoneId)}
                      className="w-4 h-4 cursor-pointer"
                    />
                  </td>
                  <td className="p-3">
                    <div className="font-medium">{camera.name}</div>
                    <div className="text-xs text-gray-500">
                      {camera.device.serialNumber}
                    </div>
                  </td>
                  <td className="p-3">
                    <div>{camera.device.ip}</div>
                    <div className="text-xs text-gray-500">
                      Port: {camera.device.port}
                    </div>
                  </td>
                  <td className="p-3">{camera.device.manufacturer}</td>
                  <td className="p-3">{camera.device.model}</td>
                  <td className="p-3 text-xs">{camera.device.firmwareVersion}</td>
                  <td className="p-3">
                    <div className="space-y-1">
                      {camera.streams.map((stream, idx) => (
                        <div key={idx} className="text-xs">
                          <span className="text-blue-400">{stream.resolution}</span>
                          <span className="text-gray-500 ml-1">
                            ({stream.encoding})
                          </span>
                        </div>
                      ))}
                    </div>
                  </td>
                  <td className="p-3">
                    {camera.ptzCapabilities.pan ||
                    camera.ptzCapabilities.tilt ||
                    camera.ptzCapabilities.zoom ? (
                      <span className="text-green-400">✓ Yes</span>
                    ) : (
                      <span className="text-gray-500">✗ No</span>
                    )}
                  </td>
                  <td className="p-3">
                    <span
                      className={`px-2 py-1 rounded text-xs ${
                        camera.status === 'ONLINE'
                          ? 'bg-green-900 text-green-200'
                          : 'bg-gray-700 text-gray-300'
                      }`}
                    >
                      {camera.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Empty State */}
      {cameras.length === 0 && !isDiscovering && (
        <div className="flex-1 flex items-center justify-center text-gray-500">
          <div className="text-center">
            <div className="text-6xl mb-4">🎥</div>
            <p>No cameras discovered yet</p>
            <p className="text-sm mt-2">Click "Discover Cameras" to start</p>
          </div>
        </div>
      )}
    </div>
  );
}
