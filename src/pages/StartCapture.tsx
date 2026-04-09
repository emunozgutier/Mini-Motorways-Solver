import React, { useRef, useEffect, useState } from 'react';
import { useCapture } from '../store/useCapture';

const StartCapture: React.FC = () => {
  const { step, stream, cropArea, setStep, setStream, setCropArea, setActivePage } = useCapture();
  const videoRef = useRef<HTMLVideoElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  
  const [isDragging, setIsDragging] = useState(false);
  const [isResizing, setIsResizing] = useState<string | null>(null);
  const [startPos, setStartPos] = useState({ x: 0, y: 0, cropX: 0, cropY: 0, cropW: 0, cropH: 0 });

  const startCapture = async () => {
    try {
      const mediaStream = await navigator.mediaDevices.getDisplayMedia({
        video: { cursor: "always" } as any,
        audio: false
      });
      setStream(mediaStream);
      setStep('WINDOW_MODE');
    } catch (err) {
      console.error("Error capturing screen:", err);
    }
  };

  useEffect(() => {
    if (stream && videoRef.current) {
      videoRef.current.srcObject = stream;
    }
  }, [stream, step]);

  const onMouseDown = (e: React.MouseEvent, type: string | null) => {
    e.preventDefault();
    e.stopPropagation(); 
    
    if (!containerRef.current) return;
    
    setStartPos({
      x: e.clientX,
      y: e.clientY,
      cropX: cropArea.x,
      cropY: cropArea.y,
      cropW: cropArea.width,
      cropH: cropArea.height
    });

    if (type === 'move') {
      setIsDragging(true);
    } else {
      setIsResizing(type);
    }
  };

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging && !isResizing) return;
      if (!containerRef.current) return;

      const rect = containerRef.current.getBoundingClientRect();
      const dx = ((e.clientX - startPos.x) / rect.width) * 100;
      const dy = ((e.clientY - startPos.y) / rect.height) * 100;

      if (isDragging) {
        setCropArea({
          ...cropArea,
          x: Math.max(0, Math.min(100 - startPos.cropW, startPos.cropX + dx)),
          y: Math.max(0, Math.min(100 - startPos.cropH, startPos.cropY + dy))
        });
      } else if (isResizing) {
        let x = startPos.cropX;
        let y = startPos.cropY;
        let width = startPos.cropW;
        let height = startPos.cropH;
        
        if (isResizing.includes('e')) width = Math.max(5, Math.min(100 - x, startPos.cropW + dx));
        if (isResizing.includes('s')) height = Math.max(5, Math.min(100 - y, startPos.cropH + dy));
        if (isResizing.includes('w')) {
          const newX = Math.max(0, Math.min(startPos.cropX + startPos.cropW - 5, startPos.cropX + dx));
          width = startPos.cropW - (newX - startPos.cropX);
          x = newX;
        }
        if (isResizing.includes('n')) {
          const newY = Math.max(0, Math.min(startPos.cropY + startPos.cropH - 5, startPos.cropY + dy));
          height = startPos.cropH - (newY - startPos.cropY);
          y = newY;
        }
        setCropArea({ x, y, width, height });
      }
    };

    const handleMouseUp = () => {
      setIsDragging(false);
      setIsResizing(null);
    };

    if (isDragging || isResizing) {
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
    }
    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging, isResizing, startPos, setCropArea, cropArea.width, cropArea.height]);

  const handleWindowMode = (windowed: boolean) => {
    if (windowed) {
      setStep('SELECT_AREA');
    } else {
      setStep('READY');
      setActivePage('DISPLAY');
    }
  };

  const renderStep = () => {
    switch (step) {
      case 'SELECT_SOURCE':
        return (
          <div className="step-card">
            <h2>Step 1: Select Display</h2>
            <p>To help you play Mini Motorways, I need to see the game. Please select the display where the game is running.</p>
            <button className="btn-primary" onClick={startCapture}>Select Display</button>
          </div>
        );
      case 'WINDOW_MODE':
        return (
          <div className="step-card">
            <h2>Step 2: Display Mode</h2>
            <p>Is the game running in full screen or in a window?</p>
            <div className="button-group">
              <button className="btn-secondary" onClick={() => handleWindowMode(false)}>Full Screen</button>
              <button className="btn-secondary" onClick={() => handleWindowMode(true)}>Windowed</button>
            </div>
          </div>
        );
      case 'SELECT_AREA':
        return (
          <div className="step-card large">
            <h2>Step 3: Define Game Area</h2>
            <p>Drag and resize the box to cover exactly the game board area.</p>
            <div className="crop-container" ref={containerRef}>
              <video ref={videoRef} autoPlay playsInline muted />
              <div className="crop-overlay">
                <div 
                  className="crop-box" 
                  onMouseDown={(e) => onMouseDown(e, 'move')}
                  style={{ 
                    left: `${cropArea.x}%`, 
                    top: `${cropArea.y}%`, 
                    width: `${cropArea.width}%`, 
                    height: `${cropArea.height}%` 
                  }}
                >
                  <div className="crop-handle nw" onMouseDown={(e) => onMouseDown(e, 'nw')} />
                  <div className="crop-handle ne" onMouseDown={(e) => onMouseDown(e, 'ne')} />
                  <div className="crop-handle sw" onMouseDown={(e) => onMouseDown(e, 'sw')} />
                  <div className="crop-handle se" onMouseDown={(e) => onMouseDown(e, 'se')} />
                  <div className="crop-handle n" onMouseDown={(e) => onMouseDown(e, 'n')} />
                  <div className="crop-handle s" onMouseDown={(e) => onMouseDown(e, 's')} />
                  <div className="crop-handle e" onMouseDown={(e) => onMouseDown(e, 'e')} />
                  <div className="crop-handle w" onMouseDown={(e) => onMouseDown(e, 'w')} />
                </div>
              </div>
            </div>
            <button className="btn-primary" onClick={() => { setStep('READY'); setActivePage('DISPLAY'); } }>Confirm Area</button>
          </div>
        );
      case 'READY':
        return (
          <div className="step-card">
            <h2>Capture Ready!</h2>
            <p>The solver is now monitoring the selected area.</p>
            <div className="stream-preview mini">
              <video ref={videoRef} autoPlay playsInline muted />
            </div>
            <button className="btn-secondary" onClick={() => setActivePage('DISPLAY')}>Go to Display</button>
          </div>
        );
    }
  };

  return (
    <div className="wizard-container">
      {renderStep()}
    </div>
  );
};

export default StartCapture;

