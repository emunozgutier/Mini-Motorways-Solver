import React, { useRef, useEffect, useState } from 'react';
import { useCapture } from '../store/useCapture';

const rgbToHsl = (r: number, g: number, b: number) => {
  r /= 255; g /= 255; b /= 255;
  const max = Math.max(r, g, b), min = Math.min(r, g, b);
  let h = 0, s, l = (max + min) / 2;

  if (max === min) {
    h = s = 0;
  } else {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    switch (max) {
      case r: h = (g - b) / d + (g < b ? 6 : 0); break;
      case g: h = (b - r) / d + 2; break;
      case b: h = (r - g) / d + 4; break;
    }
    h /= 6;
  }

  return {
    h: Math.round(h * 360),
    s: Math.round(s * 100),
    l: Math.round(l * 100)
  };
};

const KeyColors: React.FC = () => {
  const { stream, cropArea, keyColors, addKeyColor, removeKeyColor, updateKeyColorLabel } = useCapture();
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [error, setError] = useState<string | null>(null);

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

  const handlePickColor = (e: React.MouseEvent<HTMLDivElement>) => {
    const video = videoRef.current;
    const canvas = canvasRef.current;
    if (!video || !canvas) return;

    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;

    ctx.drawImage(video, 0, 0, canvas.width, canvas.height);

    const streamX = (cropArea.x / 100) * video.videoWidth + (x / rect.width) * (cropArea.width / 100) * video.videoWidth;
    const streamY = (cropArea.y / 100) * video.videoHeight + (y / rect.height) * (cropArea.height / 100) * video.videoHeight;

    const pixel = ctx.getImageData(streamX, streamY, 1, 1).data;
    const r = pixel[0];
    const g = pixel[1];
    const b = pixel[2];

    const hex = `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)}`;
    const rgb = `rgb(${r}, ${g}, ${b})`;
    const hslData = rgbToHsl(r, g, b);
    const hsl = `H: ${hslData.h}°, S: ${hslData.s}%, L: ${hslData.l}%`;

    const result = addKeyColor({ hex, rgb, hsl, h: hslData.h });
    
    if (!result.success) {
      setError(result.reason || 'Failed to add color');
      setTimeout(() => setError(null), 3000);
    }
  };

  const scaleX = 100 / cropArea.width;
  const scaleY = 100 / cropArea.height;
  const translateX = -cropArea.x;
  const translateY = -cropArea.y;

  return (
    <div className="colors-page">
      {error && <div className="error-toast">{error}</div>}
      <div className="colors-layout">
        <div className="sampler-section">
          <div className="display-card">
            <div className="status-header">
              <div className="status-badge">Eye-dropper Mode</div>
              <span className="dimensions">Click anywhere to sample color</span>
            </div>
            
            <div 
              className="video-viewport dropper-active" 
              onClick={handlePickColor}
            >
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
            <canvas ref={canvasRef} style={{ display: 'none' }} />
          </div>
        </div>

        <div className="palette-section">
          <div className="palette-card">
            <h3>Color Palette</h3>
            <div className="color-list">
              {keyColors.length === 0 && (
                <div className="empty-palette">
                  <p>No colors sampled yet.</p>
                </div>
              )}
              {keyColors.map((color) => (
                <div key={color.id} className="color-item">
                  <div 
                    className="color-swatch" 
                    style={{ backgroundColor: color.hex }}
                  />
                  <div className="color-info">
                    <input 
                      type="text" 
                      value={color.label} 
                      onChange={(e) => updateKeyColorLabel(color.id, e.target.value)}
                      className="color-label-input"
                    />
                    <div className="color-formats">
                      <span className="color-format hex">{color.hex}</span>
                      <span className="color-format rgb">{color.rgb}</span>
                      <span className="color-format hsl">{color.hsl}</span>
                    </div>
                  </div>
                  <button 
                    className="btn-icon delete" 
                    onClick={() => removeKeyColor(color.id)}
                    title="Remove color"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default KeyColors;
