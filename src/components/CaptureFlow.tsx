import React, { useState, useRef, useEffect } from 'react';

type Step = 'SELECT_SOURCE' | 'WINDOW_MODE' | 'SELECT_AREA' | 'READY';

interface CaptureFlowProps {
  onCaptureComplete: (stream: MediaStream, cropArea: { x: number, y: number, width: number, height: number } | null) => void;
}

const CaptureFlow: React.FC<CaptureFlowProps> = ({ onCaptureComplete }) => {
  const [step, setStep] = useState<Step>('SELECT_SOURCE');
  const [stream, setStream] = useState<MediaStream | null>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  
  const [crop, setCrop] = useState({ x: 10, y: 10, width: 80, height: 80 }); 
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

  const handleWindowMode = (windowed: boolean) => {
    if (windowed) {
      setStep('SELECT_AREA');
    } else {
      finish(null);
    }
  };

  const onMouseDown = (e: React.MouseEvent, type: string | null) => {
    e.preventDefault();
    if (!containerRef.current) return;
    
    setStartPos({
      x: e.clientX,
      y: e.clientY,
      cropX: crop.x,
      cropY: crop.y,
      cropW: crop.width,
      cropH: crop.height
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
        setCrop(prev => ({
          ...prev,
          x: Math.max(0, Math.min(100 - prev.width, startPos.cropX + dx)),
          y: Math.max(0, Math.min(100 - prev.height, startPos.cropY + dy))
        }));
      } else if (isResizing) {
        setCrop(prev => {
          let { x, y, width, height } = { ...prev };
          if (isResizing.includes('e')) width = Math.max(5, Math.min(100 - x, startPos.cropW + dx));
          if (isResizing.includes('s')) height = Math.max(5, Math.min(100 - y, startPos.cropH + dy));
          if (isResizing.includes('w')) {
            const newX = Math.max(0, Math.min(prev.x + prev.width - 5, startPos.cropX + dx));
            width = startPos.cropW - (newX - startPos.cropX);
            x = newX;
          }
          if (isResizing.includes('n')) {
            const newY = Math.max(0, Math.min(prev.y + prev.height - 5, startPos.cropY + dy));
            height = startPos.cropH - (newY - startPos.cropY);
            y = newY;
          }
          return { x, y, width, height };
        });
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
  }, [isDragging, isResizing, startPos]);

  const finish = (finalCrop: typeof crop | null) => {
    if (stream) {
      onCaptureComplete(stream, finalCrop);
      setStep('READY');
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
                    left: `${crop.x}%`, 
                    top: `${crop.y}%`, 
                    width: `${crop.width}%`, 
                    height: `${crop.height}%` 
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
            <button className="btn-primary" onClick={() => finish(crop)}>Confirm Area</button>
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
            <button className="btn-secondary" onClick={() => setStep('SELECT_SOURCE')}>Reset Capture</button>
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

export default CaptureFlow;
