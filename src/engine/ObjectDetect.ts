import type { KeyColor } from '../store/useCapture';

import { rgbToHsl, isHueMatch } from '../utils/colorUtils';

export interface Detection {
  colorId: string;
  label: string;
  x: number;
  y: number;
  width: number;
  height: number;
  color: string;
}

/**
 * Object Detection Engine
 * Scans a frame and identifies objects matching the hue of key colors.
 */
export class ObjectDetector {
  private offscreenCanvas: HTMLCanvasElement;
  private offscreenCtx: CanvasRenderingContext2D;

  constructor(width = 160, height = 90) {
    this.offscreenCanvas = document.createElement('canvas');
    this.offscreenCanvas.width = width;
    this.offscreenCanvas.height = height;
    this.offscreenCtx = this.offscreenCanvas.getContext('2d', { willReadFrequently: true })!;
  }

  detect(source: CanvasImageSource, keyColors: KeyColor[]): Detection[] {
    if (keyColors.length === 0) return [];

    const { width, height } = this.offscreenCanvas;
    
    // Draw source to small canvas for fast processing
    this.offscreenCtx.drawImage(source, 0, 0, width, height);
    const imageData = this.offscreenCtx.getImageData(0, 0, width, height);
    const pixels = imageData.data;

    const detections: Detection[] = [];

    // Scan for each color

    for (const keyColor of keyColors) {
      const colorDetections: { minX: number, maxX: number, minY: number, maxY: number, pixels: number }[] = [];

      for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
          const idx = (y * width + x) * 4;
          const r = pixels[idx];
          const g = pixels[idx + 1];
          const b = pixels[idx + 2];

          const hsl = rgbToHsl(r, g, b);

          if (isHueMatch(hsl.h, keyColor.h, 10)) {
            // Find if this pixel belongs to an existing detection (blob)
            let found = false;
            // Simple approach: if it's within a few pixels of an existing box, merge it
            for (const det of colorDetections) {
              if (x >= det.minX - 2 && x <= det.maxX + 2 && 
                  y >= det.minY - 2 && y <= det.maxY + 2) {
                det.minX = Math.min(det.minX, x);
                det.maxX = Math.max(det.maxX, x);
                det.minY = Math.min(det.minY, y);
                det.maxY = Math.max(det.maxY, y);
                det.pixels++;
                found = true;
                break;
              }
            }

            if (!found) {
              colorDetections.push({ minX: x, maxX: x, minY: y, maxY: y, pixels: 1 });
            }
          }
        }
      }

      // Finalize boxes for this color
      const minW = keyColor.minWidth || 0;
      const minH = keyColor.minHeight || 0;

      for (const det of colorDetections) {
        const detW = ((det.maxX - det.minX + 1) / width) * 100;
        const detH = ((det.maxY - det.minY + 1) / height) * 100;

        // Filter based on min size %
        if (det.pixels > 3 && detW >= minW && detH >= minH) {
          detections.push({
            colorId: keyColor.id,
            label: keyColor.label,
            x: (det.minX / width) * 100,
            y: (det.minY / height) * 100,
            width: detW,
            height: detH,
            color: keyColor.hex
          });
        }
      }

    }

    return detections;
  }
}
