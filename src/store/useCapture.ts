import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { hueDiff } from '../utils/colorUtils';

export type CaptureStep = 'SELECT_SOURCE' | 'WINDOW_MODE' | 'SELECT_AREA' | 'READY';

export type ActivePage = 'START' | 'DISPLAY' | 'COLORS' | 'DETECTION';

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
  s: number;
  l: number;
  label: string;
  minWidth: number;   // Percentage (0-100)
  minHeight: number;  // Percentage (0-100)
  tolHue: number;     // Degrees (0-180)
  tolSat: number;     // Percentage (0-100)
  tolLum: number;     // Percentage (0-100)
}

interface CaptureState {
  step: CaptureStep;
  activePage: ActivePage;
  stream: MediaStream | null;
  cropArea: CropArea;
  keyColors: KeyColor[];
  selectedColorId: string | null;
  setStep: (step: CaptureStep) => void;
  setActivePage: (page: ActivePage) => void;
  setStream: (stream: MediaStream | null) => void;
  setCropArea: (cropArea: CropArea) => void;
  setSelectedColorId: (id: string | null) => void;
  addKeyColor: (color: Omit<KeyColor, 'label' | 'id' | 'minWidth' | 'minHeight' | 'tolHue' | 'tolSat' | 'tolLum'>) => { success: boolean, reason?: string };
  removeKeyColor: (id: string) => void;
  updateKeyColorLabel: (id: string, label: string) => void;
  updateKeyColorConfig: (id: string, config: Partial<Pick<KeyColor, 'minWidth' | 'minHeight' | 'tolHue' | 'tolSat' | 'tolLum'>>) => void;
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
      selectedColorId: null,
      setStep: (step) => set({ step }),
      setActivePage: (activePage) => set({ activePage }),
      setStream: (stream) => set({ stream }),
      setCropArea: (cropArea) => set({ cropArea }),
      setSelectedColorId: (selectedColorId) => set({ selectedColorId }),
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
            { 
              ...color, 
              id: Math.random().toString(36).substr(2, 9), 
              label: `Color ${state.keyColors.length + 1}`,
              minWidth: 1,
              minHeight: 1,
              tolHue: 10,
              tolSat: 20,
              tolLum: 20
            }
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
      updateKeyColorConfig: (id, config) => set((state) => ({
        keyColors: state.keyColors.map((c) => c.id === id ? { ...c, ...config } : c)
      })),
      reset: () => set({ 
        step: 'SELECT_SOURCE', 
        activePage: 'START',
        stream: null, 
        cropArea: { x: 10, y: 10, width: 80, height: 80 },
        keyColors: [],
        selectedColorId: null
      }),


    }),
    {
      name: 'mini-motorways-capture-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ 
        cropArea: state.cropArea, 
        keyColors: state.keyColors,
        selectedColorId: state.selectedColorId,
        activePage: state.activePage === 'START' ? 'START' : state.activePage
      }),
    }
  )
);

