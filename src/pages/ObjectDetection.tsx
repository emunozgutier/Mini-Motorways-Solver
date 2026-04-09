import React, { useRef, useEffect, useState } from 'react';
import { useCapture } from '../store/useCapture';
import { rgbToHsl, isHueMatch } from '../utils/colorUtils';

const ObjectDetection: React.FC = () => {
  const { stream, cropArea, keyColors, selectedColorId, setSelectedColorId } = useCapture();
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [tolerance, setTolerance] = useState(10);

  const selectedColor = keyColors.find(c => c.id === selectedColorId);

  useEffect(() => {
    if (stream && videoRef.current) {
      videoRef.current.srcObject = stream;
    }
  }, [stream]);

  useEffect(() => {
    let animationFrame: number;

    const processFrame = () => {
      if (!videoRef.current || !canvasRef.current || !selectedColor) {
        animationFrame = requestAnimationFrame(processFrame);
        return;
      }

      const video = videoRef.current;
      const canvas = canvasRef.current;
      const ctx = canvas.getContext('2d', { willReadFrequently: true });
      if (!ctx || video.paused || video.ended) {
        animationFrame = requestAnimationFrame(processFrame);
        return;
      }

      // For isolation, we use a resolution that balances quality and performance
      const procWidth = 640;
      const procHeight = 360;
      canvas.width = procWidth;
      canvas.height = procHeight;

      // Draw the video frame to the processing canvas
      ctx.drawImage(video, 0, 0, procWidth, procHeight);
      
      const imageData = ctx.getImageData(0, 0, procWidth, procHeight);
      const data = imageData.data;

      for (let i = 0; i < data.length; i += 4) {
        const r = data[i];
        const g = data[i + 1];
        const b = data[i + 2];

        const hsl = rgbToHsl(r, g, b);

        if (!isHueMatch(hsl.h, selectedColor.h, tolerance)) {
          // Black out non-matching pixels
          data[i] = 0;     // R
          data[i + 1] = 0; // G
          data[i + 2] = 0; // B
          // Alpha remains 255
        }
      }

      ctx.putImageData(imageData, 0, 0);

      animationFrame = requestAnimationFrame(processFrame);
    };

    animationFrame = requestAnimationFrame(processFrame);
    return () => cancelAnimationFrame(animationFrame);
  }, [selectedColor, tolerance]);

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
    <div className="detection-page">
      <div className="detection-controls display-card">
        <div className="control-group">
          <label>Isolate Color:</label>
          <select 
            value={selectedColorId || ''} 
            onChange={(e) => setSelectedColorId(e.target.value || null)}
            className="color-select"
          >
            <option value="">Select a color...</option>
            {keyColors.map(color => (
              <option key={color.id} value={color.id}>
                {color.label} ({color.hex})
              </option>
            ))}
          </select>
        </div>

        <div className="control-group">
          <div className="label-row">
            <label>Hue Tolerance:</label>
            <span className="value">{tolerance}°</span>
          </div>
          <input 
            type="range" 
            min="1" 
            max="45" 
            value={tolerance} 
            onChange={(e) => setTolerance(parseInt(e.target.value))}
            className="range-input"
          />
        </div>
      </div>

      <div className="detection-layout">
        <div className="main-viewport isolation-mode">
          <div className="display-card">
            <div className="status-header">
              <div className="status-badge isolation">Color Isolation View</div>
              {selectedColor && (
                <div className="target-swatch">
                  <div className="swatch" style={{ backgroundColor: selectedColor.hex }} />
                  <span>Isolating {selectedColor.label}</span>
                </div>
              )}
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
                  style={{ display: 'none' }} // Hide raw video 
                />
                <canvas ref={canvasRef} className="isolation-canvas" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ObjectDetection;
