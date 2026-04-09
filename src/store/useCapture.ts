import { create } from 'zustand';

export type CaptureStep = 'SELECT_SOURCE' | 'WINDOW_MODE' | 'SELECT_AREA' | 'READY';
export type ActivePage = 'START' | 'DISPLAY';

interface CropArea {
  x: number;
  y: number;
  width: number;
  height: number;
}

interface CaptureState {
  step: CaptureStep;
  activePage: ActivePage;
  stream: MediaStream | null;
  cropArea: CropArea;
  setStep: (step: CaptureStep) => void;
  setActivePage: (page: ActivePage) => void;
  setStream: (stream: MediaStream | null) => void;
  setCropArea: (cropArea: CropArea) => void;
  reset: () => void;
}

export const useCapture = create<CaptureState>((set) => ({
  step: 'SELECT_SOURCE',
  activePage: 'START',
  stream: null,
  cropArea: { x: 10, y: 10, width: 80, height: 80 },
  setStep: (step) => set({ step }),
  setActivePage: (activePage) => set({ activePage }),
  setStream: (stream) => set({ stream }),
  setCropArea: (cropArea) => set({ cropArea }),
  reset: () => set({ 
    step: 'SELECT_SOURCE', 
    activePage: 'START',
    stream: null, 
    cropArea: { x: 10, y: 10, width: 80, height: 80 } 
  }),
}));

