package client

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
)

// DockerClient handles Docker container operations
type DockerClient struct {
	cli    *client.Client
	logger zerolog.Logger
}

// NewDockerClient creates a new Docker client
func NewDockerClient(logger zerolog.Logger) (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerClient{
		cli:    cli,
		logger: logger,
	}, nil
}

// WHIPPusherConfig contains configuration for WHIP pusher container
type WHIPPusherConfig struct {
	ContainerName string
	RTSPURL       string
	WHIPEndpoint  string
	StreamKey     string
	NetworkName   string
}

// StartWHIPPusher starts a WHIP pusher container
func (d *DockerClient) StartWHIPPusher(ctx context.Context, config WHIPPusherConfig) error {
	// Check if container already exists and remove it
	if err := d.RemoveContainer(ctx, config.ContainerName); err != nil {
		d.logger.Warn().Err(err).Str("container", config.ContainerName).Msg("Failed to remove existing container")
	}

	// Container configuration
	containerConfig := &container.Config{
		Image: "whip-pusher:latest",
		Env: []string{
			fmt.Sprintf("RTSP_URL=%s", config.RTSPURL),
			fmt.Sprintf("WHIP_ENDPOINT=%s", config.WHIPEndpoint),
			fmt.Sprintf("STREAM_KEY=%s", config.StreamKey),
		},
		Labels: map[string]string{
			"app":     "cctv-whip-pusher",
			"managed": "go-api",
		},
	}

	// Host configuration
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		NetworkMode: container.NetworkMode(config.NetworkName),
	}

	// Network configuration
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			config.NetworkName: {},
		},
	}

	// Create container
	resp, err := d.cli.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		config.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := d.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	d.logger.Info().
		Str("container_id", resp.ID).
		Str("container_name", config.ContainerName).
		Str("rtsp_url", config.RTSPURL).
		Str("whip_endpoint", config.WHIPEndpoint).
		Msg("Started WHIP pusher container")

	return nil
}

// StopWHIPPusher stops and removes a WHIP pusher container
func (d *DockerClient) StopWHIPPusher(ctx context.Context, containerName string) error {
	return d.RemoveContainer(ctx, containerName)
}

// RemoveContainer removes a container (stops it first if running)
func (d *DockerClient) RemoveContainer(ctx context.Context, containerName string) error {
	// Try to stop the container (ignore errors if not running)
	timeout := 10
	_ = d.cli.ContainerStop(ctx, containerName, container.StopOptions{Timeout: &timeout})

	// Remove the container
	err := d.cli.ContainerRemove(ctx, containerName, container.RemoveOptions{
		Force: true,
	})
	if err != nil && !client.IsErrNotFound(err) {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	d.logger.Info().
		Str("container", containerName).
		Msg("Removed WHIP pusher container")

	return nil
}

// GetContainerLogs retrieves logs from a container
func (d *DockerClient) GetContainerLogs(ctx context.Context, containerName string) (string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "50",
	}

	logs, err := d.cli.ContainerLogs(ctx, containerName, options)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return string(logBytes), nil
}

// Close closes the Docker client connection
func (d *DockerClient) Close() error {
	return d.cli.Close()
}
