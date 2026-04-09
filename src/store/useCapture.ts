import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { hueDiff } from '../utils/colorUtils';

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
  hsl: string;
  h: number; // For diffing
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
  addKeyColor: (color: Omit<KeyColor, 'label' | 'id'>) => { success: boolean, reason?: string };
  removeKeyColor: (id: string) => void;
  updateKeyColorLabel: (id: string, label: string) => void;
  reset: () => void;
}

export const useCapture = create<CaptureState>()(

  persist(
    (set, get) => ({
      step: 'SELECT_SOURCE',
      activePage: 'START',
      stream: null,
      cropArea: { x: 10, y: 10, width: 80, height: 80 },
      keyColors: [],
      setStep: (step) => set({ step }),
      setActivePage: (activePage) => set({ activePage }),
      setStream: (stream) => set({ stream }),
      setCropArea: (cropArea) => set({ cropArea }),
      addKeyColor: (color) => {
        const { keyColors } = get();
        
        // Hue check (at least 3 degrees apart)
        const isTooClose = keyColors.some(existing => hueDiff(existing.h, color.h) < 3);
        
        if (isTooClose) {
          return { success: false, reason: 'Color hue is too similar to an existing one (needs at least 3° difference).' };
        }

        set((state) => ({
          keyColors: [
            ...state.keyColors,
            { ...color, id: Math.random().toString(36).substr(2, 9), label: `Color ${state.keyColors.length + 1}` }
          ]
        }));
        return { success: true };
      },
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
    }),
    {
      name: 'mini-motorways-capture-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ 
        cropArea: state.cropArea, 
        keyColors: state.keyColors,
        activePage: state.activePage === 'START' ? 'START' : 'DISPLAY'
      }),
    }
  )
);
