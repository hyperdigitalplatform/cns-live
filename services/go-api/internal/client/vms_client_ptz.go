package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rta/cctv/go-api/internal/domain"
)

// translatePTZCommand converts simple command strings to VMS PTZ format
func (c *VMSClient) translatePTZCommand(cmd domain.PTZCommand) map[string]interface{} {
	vmsCmd := map[string]interface{}{
		"camera_id": cmd.CameraID,
		"user_id":   cmd.UserID,
		"speed":     cmd.Speed,
	}

	switch cmd.Command {
	case "pan_left":
		vmsCmd["action"] = "MOVE"
		vmsCmd["pan"] = -cmd.Speed
		vmsCmd["tilt"] = 0.0
		vmsCmd["zoom"] = 0.0

	case "pan_right":
		vmsCmd["action"] = "MOVE"
		vmsCmd["pan"] = cmd.Speed
		vmsCmd["tilt"] = 0.0
		vmsCmd["zoom"] = 0.0

	case "tilt_up":
		vmsCmd["action"] = "MOVE"
		vmsCmd["pan"] = 0.0
		vmsCmd["tilt"] = cmd.Speed
		vmsCmd["zoom"] = 0.0

	case "tilt_down":
		vmsCmd["action"] = "MOVE"
		vmsCmd["pan"] = 0.0
		vmsCmd["tilt"] = -cmd.Speed
		vmsCmd["zoom"] = 0.0

	case "zoom_in":
		vmsCmd["action"] = "MOVE"
		vmsCmd["pan"] = 0.0
		vmsCmd["tilt"] = 0.0
		vmsCmd["zoom"] = cmd.Speed

	case "zoom_out":
		vmsCmd["action"] = "MOVE"
		vmsCmd["pan"] = 0.0
		vmsCmd["tilt"] = 0.0
		vmsCmd["zoom"] = -cmd.Speed

	case "stop":
		vmsCmd["action"] = "STOP"

	case "home":
		vmsCmd["action"] = "GO_TO_PRESET"
		vmsCmd["preset"] = 1 // Preset 1 = home position

	default:
		vmsCmd["action"] = "MOVE"
		vmsCmd["pan"] = 0.0
		vmsCmd["tilt"] = 0.0
		vmsCmd["zoom"] = 0.0
	}

	return vmsCmd
}

// ControlPTZTranslated sends PTZ control command to VMS with translation
func (c *VMSClient) ControlPTZTranslated(ctx context.Context, cmd domain.PTZCommand) error {
	endpoint := fmt.Sprintf("%s/vms/cameras/%s/ptz", c.baseURL, cmd.CameraID)

	// Translate command to VMS format
	vmsCmd := c.translatePTZCommand(cmd)

	reqBody, err := json.Marshal(vmsCmd)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to control PTZ: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vms service returned status %d", resp.StatusCode)
	}

	c.logger.Info().
		Str("camera_id", cmd.CameraID).
		Str("command", cmd.Command).
		Str("action", vmsCmd["action"].(string)).
		Str("user_id", cmd.UserID).
		Msg("PTZ command executed")

	return nil
}
