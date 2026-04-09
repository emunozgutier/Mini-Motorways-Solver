import { create } from 'zustand';

export type CaptureStep = 'SELECT_SOURCE' | 'WINDOW_MODE' | 'SELECT_AREA' | 'READY';
export type ActivePage = 'START' | 'DISPLAY' | 'COLORS';

interface CropArea {
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface KeyColor {
  id: string;
  hex: string;
  rgb: string;
  label: string;
}

interface CaptureState {
  step: CaptureStep;
  activePage: ActivePage;
  stream: MediaStream | null;
  cropArea: CropArea;
  keyColors: KeyColor[];
  setStep: (step: CaptureStep) => void;
  setActivePage: (page: ActivePage) => void;
  setStream: (stream: MediaStream | null) => void;
  setCropArea: (cropArea: CropArea) => void;
  addKeyColor: (color: Omit<KeyColor, 'label' | 'id'>) => void;
  removeKeyColor: (id: string) => void;
  updateKeyColorLabel: (id: string, label: string) => void;
  reset: () => void;
}

export const useCapture = create<CaptureState>((set) => ({
  step: 'SELECT_SOURCE',
  activePage: 'START',
  stream: null,
  cropArea: { x: 10, y: 10, width: 80, height: 80 },
  keyColors: [],
  setStep: (step) => set({ step }),
  setActivePage: (activePage) => set({ activePage }),
  setStream: (stream) => set({ stream }),
  setCropArea: (cropArea) => set({ cropArea }),
  addKeyColor: (color) => set((state) => ({
    keyColors: [
      ...state.keyColors,
      { ...color, id: Math.random().toString(36).substr(2, 9), label: `Color ${state.keyColors.length + 1}` }
    ]
  })),
  removeKeyColor: (id) => set((state) => ({
    keyColors: state.keyColors.filter((c) => c.id !== id)
  })),
  updateKeyColorLabel: (id, label) => set((state) => ({
    keyColors: state.keyColors.map((c) => c.id === id ? { ...c, label } : c)
  })),
  reset: () => set({ 
    step: 'SELECT_SOURCE', 
    activePage: 'START',
    stream: null, 
    cropArea: { x: 10, y: 10, width: 80, height: 80 },
    keyColors: []
  }),
}));


