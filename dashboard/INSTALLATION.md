# Dashboard Installation & Setup Guide

## Quick Start

### 1. Install Dependencies

```bash
cd dashboard
npm install
```

This will install all required packages including:
- React 18
- TypeScript
- Zustand (state management with persist middleware)
- @dnd-kit (drag-and-drop)
- LiveKit components
- Lucide React (icons)
- Tailwind CSS

### 2. Install New Drag-and-Drop Packages

The following packages were added for the tree view and drag-and-drop features:

```bash
npm install @dnd-kit/core@^6.1.0 @dnd-kit/sortable@^8.0.0 @dnd-kit/utilities@^3.2.2
```

**Note**: These are already included in `package.json`, so `npm install` will install them automatically.

### 3. Run Development Server

```bash
npm run dev
```

Dashboard will be available at `http://localhost:5173`

### 4. Build for Production

```bash
npm run build
```

Build output will be in `dist/` directory.

---

## New Features Setup

### Component Files Created

The following new components were created for the enhanced dashboard:

#### 1. Type Definitions
- `src/types/index.ts` - Updated with folder and drag-and-drop types

#### 2. Stores (State Management)
- `src/stores/folderStore.ts` - **NEW** - Folder CRUD operations and tree building

#### 3. Components
- `src/components/CameraTreeView.tsx` - **NEW** - Tree view with drag-and-drop
- `src/components/CameraSidebarNew.tsx` - **NEW** - Enhanced sidebar with tree view
- `src/components/StreamGridEnhanced.tsx` - **NEW** - Grid with drop zones

#### 4. Pages
- `src/pages/LiveViewEnhanced.tsx` - **NEW** - Integrated live view page

### Updating Existing Code

To use the new features, update your `App.tsx`:

```typescript
// Old
import { LiveView } from '@/pages/LiveView';

// New
import { LiveViewEnhanced } from '@/pages/LiveViewEnhanced';

// In your routes
<Route path="/" element={<LiveViewEnhanced />} />
```

---

## Package Details

### New Dependencies

#### @dnd-kit/core
**Purpose**: Core drag-and-drop functionality
**Features**:
- Accessibility built-in
- Touch support
- Customizable sensors
- Performance optimized

#### @dnd-kit/sortable
**Purpose**: Sortable lists and trees
**Features**:
- Multi-container support
- Nested sorting
- Animation support

#### @dnd-kit/utilities
**Purpose**: Utility functions for dnd-kit
**Features**:
- CSS transform utilities
- Collision detection helpers

### Existing Dependencies (Already Installed)

- `zustand` - State management (already includes persist middleware)
- `lucide-react` - Icons for UI
- `clsx` + `tailwind-merge` - Conditional CSS classes
- `react-router-dom` - Navigation

---

## Environment Configuration

Create `.env` file in `dashboard/` directory:

```bash
# API Configuration
VITE_API_URL=http://localhost:8088
VITE_API_WS_URL=ws://localhost:8088

# LiveKit Configuration
VITE_LIVEKIT_URL=ws://localhost:7880
VITE_LIVEKIT_API_KEY=your-api-key
VITE_LIVEKIT_API_SECRET=your-api-secret

# Feature Flags
VITE_ENABLE_FOLDERS=true
VITE_ENABLE_DRAG_DROP=true
```

---

## Development Workflow

### 1. Start Dashboard

```bash
cd dashboard
npm run dev
```

### 2. Start Backend Services

In separate terminal:

```bash
cd ..
docker-compose up -d
```

### 3. Verify Services

```bash
# Check go-api
curl http://localhost:8088/health

# Check LiveKit
curl http://localhost:7880/
```

### 4. Access Dashboard

Open browser: `http://localhost:5173`

---

## Testing Drag-and-Drop

### Test Camera to Folder

1. Start dashboard
2. Cameras load in sidebar
3. Default folders created automatically
4. Drag camera from "Unorganized"
5. Drop onto "Dubai Police" folder
6. Camera moves to folder

### Test Camera to Grid

1. Select grid layout (e.g., 3×3)
2. Drag camera from sidebar
3. Drop onto empty grid cell
4. Stream starts playing

### Test Folder Creation

1. Click "+ Folder" button
2. Enter folder name
3. Right-click folder → Add Subfolder
4. Drag cameras into subfolders

---

## Troubleshooting

### Issue: `npm install` fails

**Solution**:
```bash
# Clear cache
npm cache clean --force

# Delete node_modules and package-lock.json
rm -rf node_modules package-lock.json

# Reinstall
npm install
```

### Issue: TypeScript errors

**Solution**:
```bash
# Ensure TypeScript version is correct
npm install typescript@^5.3.3 --save-dev

# Restart VS Code TypeScript server
# VS Code: Cmd/Ctrl + Shift + P → "TypeScript: Restart TS Server"
```

### Issue: Drag-and-drop not working

**Checklist**:
- [ ] `@dnd-kit` packages installed
- [ ] Browser supports HTML5 Drag and Drop API
- [ ] Not using mobile browser (touch requires different setup)
- [ ] Check browser console for errors

### Issue: Folders not persisting

**Solution**:
- Check browser local storage is enabled
- Not in incognito/private mode
- Clear browser cache and reload

---

## Browser Compatibility

### Supported Browsers

| Browser | Min Version | Drag-and-Drop | WebRTC Streams |
|---------|-------------|---------------|----------------|
| Chrome | 90+ | ✅ | ✅ |
| Firefox | 88+ | ✅ | ✅ |
| Safari | 14+ | ✅ | ✅ |
| Edge | 90+ | ✅ | ✅ |

### Unsupported

- Internet Explorer (all versions)
- Chrome < 90
- Firefox < 88
- Safari < 14

---

## Build Optimization

### Production Build

```bash
npm run build
```

**Output**:
- `dist/assets/*.js` - Minified JavaScript bundles
- `dist/assets/*.css` - Minified CSS
- `dist/index.html` - Entry point

### Build Size Optimization

Current bundle sizes (after gzip):
- Main bundle: ~350 KB
- Vendor bundle: ~200 KB
- **Total**: ~550 KB (target: <1 MB)

### Analyzing Bundle

```bash
npm install --save-dev rollup-plugin-visualizer
npm run build -- --mode analyze
```

Open `dist/stats.html` to see bundle composition.

---

## Docker Build

### Development

```bash
cd dashboard
docker build -t rta-cctv-dashboard:dev .
docker run -p 5173:80 rta-cctv-dashboard:dev
```

### Production

```bash
# From root directory
docker-compose build dashboard
docker-compose up -d dashboard
```

Access at `http://localhost:3000` (or configured port)

---

## Updating Dependencies

### Check for Updates

```bash
npm outdated
```

### Update Packages

```bash
# Update all packages to latest minor/patch versions
npm update

# Update specific package to latest
npm install @dnd-kit/core@latest

# Update all to latest (including major versions - be careful!)
npm install $(npm outdated | awk 'NR>1 {print $1"@latest"}')
```

### Test After Updates

```bash
npm run build
npm run preview
# Manual testing of all features
```

---

## Development Tools

### VS Code Extensions

Recommended extensions for development:

```json
{
  "recommendations": [
    "dbaeumer.vscode-eslint",
    "esbenp.prettier-vscode",
    "bradlc.vscode-tailwindcss",
    "dsznajder.es7-react-js-snippets",
    "formulahendry.auto-rename-tag"
  ]
}
```

### ESLint Configuration

Already configured in `eslintrc.json`:
- React hooks rules
- TypeScript rules
- Import ordering

### Prettier Configuration

Create `.prettierrc`:

```json
{
  "semi": true,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5",
  "printWidth": 80
}
```

---

## Performance Monitoring

### React Developer Tools

Install browser extension:
- Chrome: https://chrome.google.com/webstore (search "React Developer Tools")
- Firefox: https://addons.mozilla.org/firefox (search "React Developer Tools")

**Usage**:
1. Open DevTools
2. Click "Components" tab
3. Click "Profiler" tab
4. Record user interactions
5. Analyze render times

### Network Monitoring

Monitor API calls and WebRTC streams:
```bash
# Open DevTools → Network tab
# Filter by:
# - XHR: API requests
# - WS: WebSocket connections
# - Media: Video streams
```

---

## Continuous Integration

### GitHub Actions (Example)

```yaml
# .github/workflows/dashboard.yml
name: Dashboard CI

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '20'
      - run: cd dashboard && npm ci
      - run: cd dashboard && npm run lint
      - run: cd dashboard && npm run build
      - run: cd dashboard && npm run test
```

---

## Deployment

### Static Hosting (Netlify/Vercel)

```bash
# Build
npm run build

# Deploy to Netlify
netlify deploy --dir=dist --prod

# Deploy to Vercel
vercel --prod
```

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name dashboard.rta.ae;

    root /var/www/dashboard/dist;
    index index.html;

    # Gzip compression
    gzip on;
    gzip_types text/css application/javascript application/json;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

---

## Next Steps

After installation:

1. ✅ **Read Features Documentation**: See `DASHBOARD_FEATURES.md`
2. ✅ **Test Drag-and-Drop**: Follow test procedures above
3. ✅ **Customize Folders**: Create your organization structure
4. ✅ **Configure Grid Layouts**: Test different layouts
5. ✅ **Set Up Production**: Build and deploy

---

## Support

For issues:
- Check `DASHBOARD_FEATURES.md` for usage guide
- Check browser console for errors
- Verify all services are running (`docker-compose ps`)
- Check API connectivity (`curl http://localhost:8088/health`)

---

**Document Version**: 1.0
**Last Updated**: 2025-10-26
