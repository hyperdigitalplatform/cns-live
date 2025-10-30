package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the API routes
func SetupRouter(handler *Handler) *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Cameras
		v1.GET("/cameras", handler.GetCameras)
		v1.GET("/cameras/discover", handler.DiscoverCameras)

		// WebRTC Playback routes
		v1.POST("/cameras/:cameraId/playback/start", handler.StartWebRTCPlayback)
		v1.PUT("/playback/webrtc/answer", handler.UpdateWebRTCAnswer)
		v1.POST("/playback/webrtc/ice", handler.SendICECandidate)
		v1.GET("/playback/webrtc/ice/:sessionId", handler.GetICECandidates)
	}

	// Milestone-prefixed routes (for Kong routing)
	milestone := router.Group("/api/v1/milestone")
	{
		// Recording control
		recordings := milestone.Group("/recordings")
		{
			recordings.POST("/start", handler.StartRecording)
			recordings.POST("/stop", handler.StopRecording)
			recordings.GET("/status/:cameraId", handler.GetRecordingStatus)
		}

		// Sequences
		sequences := milestone.Group("/sequences")
		{
			sequences.GET("/types/:cameraId", handler.GetSequenceTypes)
			sequences.POST("", handler.GetSequences)
		}

		// Timeline
		milestone.POST("/timeline", handler.GetTimeline)
	}

	return router
}
