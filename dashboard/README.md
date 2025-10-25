# RTA CCTV Dashboard

Modern React-based web dashboard for the RTA CCTV Video Management System. Built with React 18, TypeScript, Vite, Tailwind CSS, LiveKit, and HLS.js.

## Features

- ✅ **Live Streaming**: Multi-camera grid layout with LiveKit WebRTC
- ✅ **Playback**: HLS-based playback with timeline controls
- ✅ **Camera Management**: Sidebar with filtering and search
- ✅ **Grid Layouts**: Multiple layouts (1×1, 2×2, 3×3, 4×4, 2×3, 3×4)
- ✅ **Fullscreen Mode**: Individual camera fullscreen viewing
- ✅ **State Management**: Zustand for efficient state handling
- ✅ **Responsive Design**: Tailwind CSS with mobile support
- ✅ **Real-time Updates**: WebSocket integration for live stats

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| React | 18.2.0 | UI framework |
| TypeScript | 5.3.3 | Type safety |
| Vite | 5.0.8 | Build tool |
| Tailwind CSS | 3.4.0 | Styling |
| LiveKit | 2.0.0 | WebRTC streaming |
| HLS.js | 1.4.14 | Video playback |
| Zustand | 4.4.7 | State management |
| React Router | 6.20.1 | Routing |

## Project Structure

```
dashboard/
├── src/
│   ├── components/         # React components
│   │   ├── LiveStreamPlayer.tsx    # LiveKit player
│   │   ├── StreamGrid.tsx          # Multi-camera grid
│   │   ├── CameraSidebar.tsx       # Camera management
│   │   └── PlaybackPlayer.tsx      # HLS playback
│   ├── pages/             # Page components
│   │   ├── LiveView.tsx   # Live streaming page
│   │   └── PlaybackView.tsx        # Playback page
│   ├── stores/            # Zustand stores
│   │   ├── cameraStore.ts # Camera state
│   │   └── streamStore.ts # Stream state
│   ├── services/          # API clients
│   │   └── api.ts         # REST API client
│   ├── types/             # TypeScript types
│   │   └── index.ts       # Shared types
│   ├── utils/             # Utility functions
│   │   └── cn.ts          # className utility
│   ├── App.tsx            # Main app component
│   ├── main.tsx           # Entry point
│   └── index.css          # Global styles
├── public/                # Static assets
├── package.json           # Dependencies
├── vite.config.ts         # Vite configuration
├── tailwind.config.js     # Tailwind configuration
├── tsconfig.json          # TypeScript configuration
├── Dockerfile             # Docker build
└── nginx.conf             # Nginx configuration
```

## Development

### Prerequisites

- Node.js 18+
- npm or yarn

### Setup

```bash
cd dashboard

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Environment Variables

Create a `.env` file:

```bash
VITE_API_URL=http://localhost:8088
```

## Components

### LiveStreamPlayer

LiveKit-based WebRTC player for live camera streams.

**Props**:
- `camera: Camera` - Camera object
- `quality?: 'high' | 'medium' | 'low'` - Stream quality
- `onError?: (error: Error) => void` - Error callback

**Features**:
- Auto stream reservation
- Heartbeat mechanism (25s interval)
- Connection status overlay
- Automatic cleanup on unmount

**Usage**:
```tsx
<LiveStreamPlayer
  camera={camera}
  quality="medium"
  onError={(err) => console.error(err)}
/>
```

### StreamGrid

Multi-camera grid layout with flexible configurations.

**Props**:
- `cameras: Camera[]` - Array of cameras to display
- `defaultLayout?: GridLayoutType` - Initial layout (default: '2x2')

**Layouts**:
- 1×1: Single camera fullscreen
- 2×2: 4 cameras
- 3×3: 9 cameras
- 4×4: 16 cameras
- 2×3: 6 cameras
- 3×4: 12 cameras

**Features**:
- Dynamic layout switching
- Individual camera fullscreen
- Empty slot placeholders
- Responsive grid

**Usage**:
```tsx
<StreamGrid
  cameras={selectedCameras}
  defaultLayout="2x2"
/>
```

### CameraSidebar

Camera management sidebar with filtering and selection.

**Props**:
- `onCameraSelect: (cameras: Camera[]) => void` - Selection callback
- `selectedCameras: Camera[]` - Currently selected cameras

**Features**:
- Search by name (English/Arabic)
- Filter by source (agency)
- Filter by status (ONLINE, OFFLINE, etc.)
- Multi-select support
- Online/offline indicators

**Usage**:
```tsx
<CameraSidebar
  onCameraSelect={setSelectedCameras}
  selectedCameras={selectedCameras}
/>
```

### PlaybackPlayer

HLS.js-based player for recorded video playback.

**Props**:
- `camera: Camera` - Camera object
- `startTime: Date` - Playback start time
- `endTime: Date` - Playback end time
- `onClose?: () => void` - Close callback

**Features**:
- HLS playback with hls.js
- Custom video controls
- Timeline scrubbing
- Fullscreen support
- Mute/unmute
- Time display

**Usage**:
```tsx
<PlaybackPlayer
  camera={camera}
  startTime={new Date('2024-01-20T10:00:00Z')}
  endTime={new Date('2024-01-20T11:00:00Z')}
  onClose={() => setShowPlayer(false)}
/>
```

## State Management

### Camera Store

Manages camera list and selection state.

```tsx
import { useCameraStore } from '@/stores/cameraStore';

function MyCom

ponent() {
  const cameras = useCameraStore(state => state.cameras);
  const fetchCameras = useCameraStore(state => state.fetchCameras);
  const selectCamera = useCameraStore(state => state.selectCamera);

  useEffect(() => {
    fetchCameras();
  }, []);

  return (
    <div>
      {cameras.map(camera => (
        <button key={camera.id} onClick={() => selectCamera(camera)}>
          {camera.name}
        </button>
      ))}
    </div>
  );
}
```

### Stream Store

Manages stream reservations and heartbeats.

```tsx
import { useStreamStore } from '@/stores/streamStore';

function MyComponent() {
  const reserveStream = useStreamStore(state => state.reserveStream);
  const releaseStream = useStreamStore(state => state.releaseStream);

  const handleWatch = async (cameraId: string) => {
    const reservation = await reserveStream(cameraId, 'medium');
    // Heartbeat started automatically
  };

  const handleStop = (reservationId: string) => {
    releaseStream(reservationId);
    // Heartbeat stopped automatically
  };

  return (
    <button onClick={() => handleWatch('cam-123')}>
      Watch Camera
    </button>
  );
}
```

## API Integration

### API Client

```tsx
import { api } from '@/services/api';

// Get cameras
const { cameras } = await api.getCameras({
  source: 'DUBAI_POLICE',
  status: 'ONLINE',
  limit: 100
});

// Reserve stream
const reservation = await api.reserveStream('camera-id', 'medium');

// Request playback
const playback = await api.requestPlayback({
  camera_id: 'camera-id',
  start_time: '2024-01-20T10:00:00Z',
  end_time: '2024-01-20T11:00:00Z',
  format: 'hls'
});
```

## Docker Deployment

### Build

```bash
docker build -t cctv-dashboard .
```

### Run

```bash
docker run -p 3000:80 \
  -e VITE_API_URL=http://localhost:8088 \
  cctv-dashboard
```

### Docker Compose

```yaml
dashboard:
  build:
    context: ./dashboard
  ports:
    - "3000:80"
  environment:
    VITE_API_URL: http://localhost:8088
  depends_on:
    - go-api
```

## Navigation

- **Live View** (`/`) - Multi-camera live streaming
- **Playback** (`/playback`) - Recorded video playback
- **Analytics** (`/analytics`) - Object detection (TODO)
- **Settings** (`/settings`) - Configuration (TODO)

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `F` | Toggle fullscreen (when camera selected) |
| `Space` | Play/Pause (in playback mode) |
| `M` | Mute/Unmute (in playback mode) |
| `←` / `→` | Seek backward/forward (in playback mode) |

## Performance

### Optimizations

- **Code Splitting**: Dynamic imports for route-based splitting
- **Lazy Loading**: Components loaded on demand
- **Memoization**: React.memo for expensive components
- **Zustand**: Minimal re-renders with selective subscriptions
- **Tailwind CSS**: Purged in production (~10 KB)

### Bundle Size

- **Main Bundle**: ~150 KB (gzipped)
- **LiveKit**: ~200 KB (gzipped)
- **HLS.js**: ~50 KB (gzipped)
- **Total**: ~400 KB (gzipped)

## Browser Support

| Browser | Version |
|---------|---------|
| Chrome | 90+ |
| Firefox | 88+ |
| Safari | 14+ |
| Edge | 90+ |

**Requirements**:
- WebRTC support (for live streaming)
- MSE support (for HLS playback)

## Troubleshooting

### Live stream not connecting

**Issue**: LiveKit connection fails

**Solutions**:
1. Check Go API is running (`http://localhost:8088`)
2. Verify LiveKit server is running
3. Check browser console for WebRTC errors
4. Ensure camera is ONLINE

### Playback not working

**Issue**: HLS playback fails

**Solutions**:
1. Check playback service is running (`http://localhost:8090`)
2. Verify recordings exist for time range
3. Check browser supports MSE (Media Source Extensions)
4. Try different browser

### Camera sidebar empty

**Issue**: No cameras shown

**Solutions**:
1. Check VMS service is running
2. Verify cameras are registered in VMS
3. Check API URL in `.env`
4. Open browser console for errors

## Future Enhancements

- [ ] JWT authentication
- [ ] User preferences persistence
- [ ] PTZ control UI
- [ ] Timeline with events
- [ ] Object detection overlays (YOLOv8)
- [ ] Export video clips
- [ ] Multi-language support (Arabic)
- [ ] Dark mode
- [ ] Mobile app (React Native)

## License

© 2024 RTA. All rights reserved.
