# PTZ Implementation Plan

**Status**: In Progress
**Last Updated**: 2025-10-27

---

## Overview

Implement PTZ (Pan-Tilt-Zoom) control for IP cameras using ONVIF protocol as primary method with RTSP control URLs as fallback.

---

## Architecture

```
Frontend (PTZControls.tsx)
  ↓ Sends: { command: "pan_left", speed: 0.5 }
  ↓
Go-API (camera_handler.go)
  ↓ Validates + Translates to ONVIF format
  ↓ Sends: { action: "MOVE", pan: -0.5, tilt: 0, zoom: 0 }
  ↓
VMS Service (milestone_repository.go)
  ↓ Tries ONVIF first
  ↓ Falls back to RTSP if ONVIF fails
  ↓
Camera (192.168.1.x)
  ✓ Executes PTZ command
```

---

## Implementation Steps

### ✅ Step 1: Enable PTZ for Camera 192.168.1.13

**Status**: COMPLETED

**Changes**:
- Updated `services/vms-service/migrations/001_create_cameras_table.sql` line 63
- Changed `ptz_enabled` from `false` to `true`
- Updated database directly: `UPDATE cameras SET ptz_enabled = true WHERE id = 'cam-002-metro-station'`

**Result**: Both cameras now show PTZ controls in dashboard

---

### 🔄 Step 2: Add Command Translation Layer

**Status**: IN PROGRESS

**File**: `services/go-api/internal/client/vms_client.go`

**Add translation function**:

```go
// translatePTZCommand converts simple command strings to VMS PTZ format
func (c *VMSClient) translatePTZCommand(cmd domain.PTZCommand) map[string]interface{} {
    vmsCmd := map[string]interface{}{
        "camera_id": cmd.CameraID,
        "user_id":   cmd.UserID,
    }

    switch cmd.Command {
    case "pan_left":
        vmsCmd["action"] = "MOVE"
        vmsCmd["pan"] = -cmd.Speed
        vmsCmd["tilt"] = 0.0
        vmsCmd["zoom"] = 0.0
        vmsCmd["speed"] = cmd.Speed

    case "pan_right":
        vmsCmd["action"] = "MOVE"
        vmsCmd["pan"] = cmd.Speed
        vmsCmd["tilt"] = 0.0
        vmsCmd["zoom"] = 0.0
        vmsCmd["speed"] = cmd.Speed

    case "tilt_up":
        vmsCmd["action"] = "MOVE"
        vmsCmd["pan"] = 0.0
        vmsCmd["tilt"] = cmd.Speed
        vmsCmd["zoom"] = 0.0
        vmsCmd["speed"] = cmd.Speed

    case "tilt_down":
        vmsCmd["action"] = "MOVE"
        vmsCmd["pan"] = 0.0
        vmsCmd["tilt"] = -cmd.Speed
        vmsCmd["zoom"] = 0.0
        vmsCmd["speed"] = cmd.Speed

    case "zoom_in":
        vmsCmd["action"] = "MOVE"
        vmsCmd["pan"] = 0.0
        vmsCmd["tilt"] = 0.0
        vmsCmd["zoom"] = cmd.Speed
        vmsCmd["speed"] = cmd.Speed

    case "zoom_out":
        vmsCmd["action"] = "MOVE"
        vmsCmd["pan"] = 0.0
        vmsCmd["tilt"] = 0.0
        vmsCmd["zoom"] = -cmd.Speed
        vmsCmd["speed"] = cmd.Speed

    case "stop":
        vmsCmd["action"] = "STOP"

    case "home":
        vmsCmd["action"] = "GO_TO_PRESET"
        vmsCmd["preset"] = 1  // Preset 1 = home position

    default:
        vmsCmd["action"] = "MOVE"
        vmsCmd["pan"] = 0.0
        vmsCmd["tilt"] = 0.0
        vmsCmd["zoom"] = 0.0
    }

    return vmsCmd
}
```

**Update ControlPTZ**:

```go
func (c *VMSClient) ControlPTZ(ctx context.Context, cmd domain.PTZCommand) error {
    endpoint := fmt.Sprintf("%s/vms/cameras/%s/ptz", c.baseURL, cmd.CameraID)

    // Translate command
    vmsCmd := c.translatePTZCommand(cmd)

    reqBody, err := json.Marshal(vmsCmd)
    if err != nil {
        return fmt.Errorf("failed to marshal request: %w", err)
    }

    // ... rest of implementation
}
```

---

### 🔄 Step 3: Implement ONVIF PTZ Control

**Status**: PENDING

**File**: `services/vms-service/internal/repository/milestone/onvif_client.go` (NEW)

**Dependencies**:
```bash
go get github.com/use-go/onvif
```

**Implementation**:

```go
package milestone

import (
    "context"
    "fmt"
    "net/url"

    "github.com/use-go/onvif"
    "github.com/use-go/onvif/ptz"
)

type ONVIFClient struct {
    device *onvif.Device
}

func NewONVIFClient(rtspURL string) (*ONVIFClient, error) {
    // Parse RTSP URL to get camera IP and credentials
    u, err := url.Parse(rtspURL)
    if err != nil {
        return nil, err
    }

    username := ""
    password := ""
    if u.User != nil {
        username = u.User.Username()
        password, _ = u.User.Password()
    }

    // Create ONVIF device
    device, err := onvif.NewDevice(onvif.DeviceParams{
        Xaddr:    fmt.Sprintf("http://%s/onvif/device_service", u.Host),
        Username: username,
        Password: password,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create ONVIF device: %w", err)
    }

    return &ONVIFClient{device: device}, nil
}

func (c *ONVIFClient) ContinuousMove(pan, tilt, zoom float64) error {
    request := ptz.ContinuousMove{
        ProfileToken: "profile_1", // Usually profile_1, may need to query
        Velocity: ptz.PTZSpeed{
            PanTilt: ptz.Vector2D{
                X: pan,
                Y: tilt,
            },
            Zoom: ptz.Vector1D{
                X: zoom,
            },
        },
    }

    _, err := c.device.CallMethod(request)
    return err
}

func (c *ONVIFClient) Stop() error {
    request := ptz.Stop{
        ProfileToken: "profile_1",
        PanTilt:      true,
        Zoom:         true,
    }

    _, err := c.device.CallMethod(request)
    return err
}

func (c *ONVIFClient) GotoHomePosition() error {
    request := ptz.GotoHomePosition{
        ProfileToken: "profile_1",
    }

    _, err := c.device.CallMethod(request)
    return err
}

func (c *ONVIFClient) GotoPreset(presetToken string) error {
    request := ptz.GotoPreset{
        ProfileToken: "profile_1",
        PresetToken:  presetToken,
    }

    _, err := c.device.CallMethod(request)
    return err
}
```

---

### 🔄 Step 4: Update Milestone Repository to Use ONVIF

**Status**: PENDING

**File**: `services/vms-service/internal/repository/milestone/milestone_repository.go`

**Add ONVIF client cache**:

```go
type MilestoneRepository struct {
    db            *sql.DB
    logger        zerolog.Logger
    onvifClients  map[string]*ONVIFClient  // Cache ONVIF clients by camera ID
    clientsMutex  sync.RWMutex
}
```

**Update ExecutePTZCommand**:

```go
func (r *MilestoneRepository) ExecutePTZCommand(ctx context.Context, cmd *domain.PTZCommand) error {
    camera, err := r.GetByID(ctx, cmd.CameraID)
    if err != nil {
        return err
    }

    if !camera.PTZEnabled {
        return fmt.Errorf("camera does not support PTZ: %s", cmd.CameraID)
    }

    // Try ONVIF first
    err = r.executePTZViaONVIF(camera, cmd)
    if err == nil {
        r.logger.Info().
            Str("camera_id", cmd.CameraID).
            Str("action", string(cmd.Action)).
            Msg("PTZ command executed via ONVIF")
        return nil
    }

    r.logger.Warn().
        Err(err).
        Str("camera_id", cmd.CameraID).
        Msg("ONVIF failed, trying RTSP fallback")

    // Fallback to RTSP
    err = r.executePTZViaRTSP(camera, cmd)
    if err != nil {
        return fmt.Errorf("PTZ command failed (ONVIF and RTSP): %w", err)
    }

    r.logger.Info().
        Str("camera_id", cmd.CameraID).
        Str("action", string(cmd.Action)).
        Msg("PTZ command executed via RTSP")

    return nil
}

func (r *MilestoneRepository) executePTZViaONVIF(camera *domain.Camera, cmd *domain.PTZCommand) error {
    // Get or create ONVIF client
    r.clientsMutex.Lock()
    client, exists := r.onvifClients[camera.ID]
    if !exists {
        var err error
        client, err = NewONVIFClient(camera.RTSPURL)
        if err != nil {
            r.clientsMutex.Unlock()
            return fmt.Errorf("failed to create ONVIF client: %w", err)
        }
        r.onvifClients[camera.ID] = client
    }
    r.clientsMutex.Unlock()

    // Execute command
    switch cmd.Action {
    case domain.PTZActionMove:
        return client.ContinuousMove(cmd.Pan, cmd.Tilt, cmd.Zoom)

    case domain.PTZActionStop:
        return client.Stop()

    case domain.PTZActionGoToPreset:
        if cmd.Preset == 1 {
            return client.GotoHomePosition()
        }
        return client.GotoPreset(fmt.Sprintf("preset_%d", cmd.Preset))

    default:
        return fmt.Errorf("unsupported PTZ action: %s", cmd.Action)
    }
}
```

---

### 🔄 Step 5: Implement RTSP Fallback

**Status**: PENDING

**File**: `services/vms-service/internal/repository/milestone/milestone_repository.go`

**Implementation**:

```go
func (r *MilestoneRepository) executePTZViaRTSP(camera *domain.Camera, cmd *domain.PTZCommand) error {
    // Parse RTSP URL
    u, err := url.Parse(camera.RTSPURL)
    if err != nil {
        return err
    }

    // Build PTZ control URL (camera-specific, may need customization)
    var ptzURL string

    switch cmd.Action {
    case domain.PTZActionMove:
        // Example for common IP cameras
        ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=start", u.Host)

        if cmd.Pan < 0 {
            ptzURL += "&direction=left"
        } else if cmd.Pan > 0 {
            ptzURL += "&direction=right"
        } else if cmd.Tilt > 0 {
            ptzURL += "&direction=up"
        } else if cmd.Tilt < 0 {
            ptzURL += "&direction=down"
        } else if cmd.Zoom > 0 {
            ptzURL += "&direction=zoomin"
        } else if cmd.Zoom < 0 {
            ptzURL += "&direction=zoomout"
        }

        ptzURL += fmt.Sprintf("&speed=%d", int(cmd.Speed*100))

    case domain.PTZActionStop:
        ptzURL = fmt.Sprintf("http://%s/cgi-bin/ptz.cgi?action=stop", u.Host)

    default:
        return fmt.Errorf("RTSP fallback not supported for action: %s", cmd.Action)
    }

    // Send HTTP request
    req, err := http.NewRequest("GET", ptzURL, nil)
    if err != nil {
        return err
    }

    // Add authentication
    if u.User != nil {
        username := u.User.Username()
        password, _ := u.User.Password()
        req.SetBasicAuth(username, password)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("camera returned status %d", resp.StatusCode)
    }

    return nil
}
```

---

### 🔄 Step 6: Add STOP Command Support

**Status**: PENDING

**Frontend**: `dashboard/src/components/PTZControls.tsx`

```typescript
const handleMouseUp = () => {
  setActiveButton(null);
  // Send STOP command
  if (activeButton) {
    handlePTZCommand('stop', { speed: 0 });
  }
};
```

**Backend**: `services/go-api/internal/delivery/http/camera_handler.go`

Add "stop" to validCommands map:

```go
validCommands := map[string]bool{
    "pan_left":  true,
    "pan_right": true,
    "tilt_up":   true,
    "tilt_down": true,
    "zoom_in":   true,
    "zoom_out":  true,
    "preset":    true,
    "home":      true,
    "stop":      true,  // ← ADD THIS
}
```

---

## Testing Plan

### Unit Tests

```go
// services/vms-service/internal/repository/milestone/onvif_client_test.go
func TestONVIFClient_ContinuousMove(t *testing.T) {
    // Mock ONVIF device
    // Test pan left/right
    // Test tilt up/down
    // Test zoom in/out
}
```

### Integration Tests

1. **Test with real camera 192.168.1.8**:
   - Send pan_left command
   - Verify camera moves left
   - Send stop command
   - Verify camera stops

2. **Test ONVIF fallback**:
   - Disable ONVIF on test camera
   - Send PTZ command
   - Verify RTSP fallback works

3. **Test frontend**:
   - Open PTZ controls
   - Press and hold pan left button
   - Release button
   - Verify camera stops

---

## Camera-Specific Configuration

Different camera brands may require different RTSP control URLs:

### Hikvision
```
http://camera-ip/ISAPI/PTZCtrl/channels/1/continuous
```

### Dahua
```
http://camera-ip/cgi-bin/ptz.cgi?action=start&channel=1&code=Left&arg1=0&arg2=5&arg3=0
```

### Generic ONVIF
```
Most modern IP cameras support ONVIF - try this first
```

**Recommendation**: Add camera brand detection and brand-specific RTSP URLs in camera metadata.

---

## Security Considerations

1. **Credentials in RTSP URL**: Currently stored in plain text
   - **TODO**: Encrypt credentials or use secrets management

2. **PTZ Command Authorization**:
   - Currently anyone with API access can control PTZ
   - **TODO**: Add RBAC (Role-Based Access Control)

3. **PTZ Command Rate Limiting**:
   - Prevent abuse/rapid commands
   - **TODO**: Add rate limiting per user/camera

---

## Performance Optimization

1. **ONVIF Client Pooling**: Cache ONVIF clients per camera (already implemented in plan)
2. **Connection Timeout**: Set reasonable timeouts (5 seconds recommended)
3. **Async Commands**: PTZ commands can be fire-and-forget (don't wait for camera response)

---

## Next Steps

1. ✅ Enable PTZ for camera 192.168.1.13 - DONE
2. ✅ Implement command translation in go-api - DONE
3. ✅ Add ONVIF client library to vms-service - DONE
4. ✅ Implement ONVIF PTZ control - DONE
5. ✅ Add RTSP fallback - DONE
6. ✅ Add STOP command support - DONE
7. ✅ Implement Milestone SDK placeholder - DONE
8. ✅ Test with real cameras - WORKING (TrueView camera at 192.168.1.13)
9. ⏳ Add configuration for method selection
10. ⏳ Add error handling and user feedback

---

## Testing Results

### Camera 1: TrueView (192.168.1.13:8888)

**Date**: 2025-10-27

**Successful Configuration**:
- ONVIF Port: 8888
- ONVIF Endpoint: `/onvif/ptz_service`
- Profile Token: `PROFILE_000` (discovered via GetProfiles)
- Authentication: WS-Security UsernameToken with digest
- Credentials: admin:pass

**CLI Testing**:
```bash
# Direct ONVIF SOAP request - SUCCESS
wget --post-file=ptz_request.xml --header='Content-Type: application/soap+xml' \
  -O - 'http://192.168.1.13:8888/onvif/ptz_service'
Response: <tptz:ContinuousMoveResponse />
```

**VMS Service Testing**:
```bash
# Through VMS service - SUCCESS
POST http://vms-service:8081/vms/cameras/cam-002-metro-station/ptz
Response: {"action":"MOVE","camera_id":"cam-002-metro-station","status":"success"}
Method: ONVIF-SOAP
```

### Camera 2: TP-Link Tapo (192.168.1.8:2020)

**Date**: 2025-10-27

**Successful Configuration**:
- ONVIF Port: 2020 (Tapo-specific)
- ONVIF Endpoint: `/onvif/service` (unified endpoint)
- Profile Token: `profile_1` (lowercase, discovered via GetProfiles)
- Authentication: WS-Security UsernameToken with digest
- Credentials: raammohan:Ilove123

**CLI Testing**:
```bash
# Direct ONVIF SOAP request - SUCCESS
wget --post-file=ptz_request.xml --header='Content-Type: application/soap+xml' \
  -O - 'http://192.168.1.8:2020/onvif/service'
Response: <tptz:ContinuousMoveResponse></tptz:ContinuousMoveResponse>
```

**VMS Service Testing**:
```bash
# Through VMS service - SUCCESS
POST http://vms-service:8081/vms/cameras/cam-001-sheikh-zayed/ptz
Response: {"action":"MOVE","camera_id":"cam-001-sheikh-zayed","status":"success"}
Method: ONVIF-SOAP
```

**Key Findings**:
1. Different camera brands use different profile tokens (PROFILE_000 vs profile_1)
2. Different endpoints: `/onvif/ptz_service` (TrueView) vs `/onvif/service` (Tapo)
3. Different ports: 8888 (TrueView) vs 2020 (Tapo)
4. WS-Security requires 20-byte nonce (not 16)
5. No HTTP Basic Auth needed (WS-Security in SOAP envelope only)
6. Fresh timestamp required (camera rejects old timestamps)
7. Auto-discovery works: code tries multiple tokens/endpoints/ports

---

**Document Version**: 4.0
**Status**: PTZ Working - Tested with Both TrueView & TP-Link Tapo Cameras
**Last Updated**: 2025-10-27
**Next Update**: After dashboard UI testing with both cameras
