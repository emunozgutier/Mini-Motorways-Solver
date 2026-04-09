import React, { useRef, useEffect } from 'react';
import { useCapture } from '../store/useCapture';

const CaptureDisplay: React.FC = () => {
  const { stream, cropArea } = useCapture();
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (stream && videoRef.current) {
      videoRef.current.srcObject = stream;
    }
  }, [stream]);

  if (!stream) {
    return (
      <div className="step-card">
        <h2>No Active Capture</h2>
        <p>Please go to the Setup tab to start a screen capture.</p>
      </div>
    );
  }

  // Calculate the scaling and positioning for the cropped view
  const scaleX = 100 / cropArea.width;
  const scaleY = 100 / cropArea.height;
  const translateX = -cropArea.x;
  const translateY = -cropArea.y;


  return (
    <div className="display-container">
      <div className="display-card">
        <div className="status-header">
          <div className="status-badge">Live Capture</div>
          <span className="dimensions">{Math.round(cropArea.width)}% x {Math.round(cropArea.height)}% area</span>
        </div>
        
        <div className="video-viewport">
          <video 
            ref={videoRef} 
            autoPlay 
            playsInline 
            muted 
            style={{
              width: `${scaleX * 100}%`,
              height: `${scaleY * 100}%`,
              transform: `translate(${translateX}%, ${translateY}%)`,
              transformOrigin: '0 0'
            }}
          />
        </div>
      </div>
    </div>
  );
};

export default CaptureDisplay;
