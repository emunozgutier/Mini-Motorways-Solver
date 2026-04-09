import React, { useEffect, useRef, useState } from 'react';
import { useCapture } from '../store/useCapture';
import { ObjectDetector } from '../engine/ObjectDetect';

import type { Detection } from '../engine/ObjectDetect';


const CaptureDisplay: React.FC = () => {
  const { stream, cropArea, keyColors } = useCapture();
  const videoRef = useRef<HTMLVideoElement>(null);
  const overlayRef = useRef<HTMLCanvasElement>(null);
  const [detections, setDetections] = useState<Detection[]>([]);
  const [isDetecting, setIsDetecting] = useState(true);
  const detectorRef = useRef<ObjectDetector>(new ObjectDetector(200, 112));

  useEffect(() => {
    if (stream && videoRef.current) {
      videoRef.current.srcObject = stream;
    }
  }, [stream]);

  useEffect(() => {
    let animationFrame: number;

    const processFrame = () => {
      if (!isDetecting || !videoRef.current || !overlayRef.current) {
        animationFrame = requestAnimationFrame(processFrame);
        return;
      }

      const video = videoRef.current;
      const overlay = overlayRef.current;
      const ctx = overlay.getContext('2d');
      if (!ctx || video.paused || video.ended) {
        animationFrame = requestAnimationFrame(processFrame);
        return;
      }

      // Detection
      const currentDetections = detectorRef.current.detect(video, keyColors);
      setDetections(currentDetections);

      // Rendering boxes to overlay
      overlay.width = video.clientWidth;
      overlay.height = video.clientHeight;
      ctx.clearRect(0, 0, overlay.width, overlay.height);

      for (const det of currentDetections) {
        const x = (det.x / 100) * overlay.width;
        const y = (det.y / 100) * overlay.height;
        const w = (det.width / 100) * overlay.width;
        const h = (det.height / 100) * overlay.height;

        ctx.strokeStyle = det.color;
        ctx.lineWidth = 3;
        ctx.strokeRect(x, y, w, h);
        
        ctx.fillStyle = det.color;
        ctx.font = 'bold 12px Inter, sans-serif';
        ctx.fillText(`${det.label}`, x, y > 15 ? y - 5 : y + 15);
      }

      animationFrame = requestAnimationFrame(processFrame);
    };

    animationFrame = requestAnimationFrame(processFrame);
    return () => cancelAnimationFrame(animationFrame);
  }, [isDetecting, keyColors]);

  if (!stream) {
    return (
      <div className="step-card">
        <h2>No Active Capture</h2>
        <p>Please go to the Setup tab to start a screen capture.</p>
      </div>
    );
  }

  const scaleX = 100 / cropArea.width;
  const scaleY = 100 / cropArea.height;
  const translateX = -cropArea.x;
  const translateY = -cropArea.y;

  return (
    <div className="display-page">
      <div className="display-layout">
        <div className="main-viewport">
          <div className="display-card">
            <div className="status-header">
              <div className="status-badge live">Live Monitoring</div>
              <div className="detection-toggle">
                <span>Detection</span>
                <button 
                  className={`toggle-btn ${isDetecting ? 'active' : ''}`}
                  onClick={() => setIsDetecting(!isDetecting)}
                >
                  {isDetecting ? 'ON' : 'OFF'}
                </button>
              </div>
            </div>
            
            <div className="video-viewport">
              <div className="video-wrapper" style={{
                width: `${scaleX * 100}%`,
                height: `${scaleY * 100}%`,
                transform: `translate(${translateX}%, ${translateY}%)`,
                transformOrigin: '0 0'
              }}>
                <video 
                  ref={videoRef} 
                  autoPlay 
                  playsInline 
                  muted 
                />
                <canvas ref={overlayRef} className="detection-overlay" />
              </div>
            </div>
          </div>
        </div>

        <div className="detection-panel">
          <div className="panel-card">
            <div className="panel-header">
              <h3>Active Objects</h3>
              <span className="count-badge">{detections.length}</span>
            </div>
            
            <div className="detection-list">
              {detections.length === 0 ? (
                <div className="empty-state">
                  <p>{isDetecting ? 'Scanning for objects...' : 'Detection is disabled'}</p>
                </div>
              ) : (
                detections.map((det, index) => (
                  <div key={index} className="detection-item" style={{ borderLeftColor: det.color }}>
                    <div className="item-main">
                      <span className="item-label">{det.label}</span>
                      <span className="item-coords">X: {Math.round(det.x)}% Y: {Math.round(det.y)}%</span>
                    </div>
                    <div className="item-dims">
                      {Math.round(det.width)}% wide × {Math.round(det.height)}% high
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CaptureDisplay;

