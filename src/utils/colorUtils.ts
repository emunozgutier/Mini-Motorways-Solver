export interface HSL {
  h: number;
  s: number;
  l: number;
}

export const rgbToHsl = (r: number, g: number, b: number): HSL => {
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

export const hueDiff = (h1: number, h2: number): number => {
  const diff = Math.abs(h1 - h2) % 360;
  return diff > 180 ? 360 - diff : diff;
};

export const isHueMatch = (h1: number, h2: number, tolerance = 10): boolean => {
  return hueDiff(h1, h2) <= tolerance;
};

export const isColorMatch = (
  c1: HSL, 
  c2: HSL, 
  tols: { h: number, s: number, l: number }
): boolean => {
  const hMatch = hueDiff(c1.h, c2.h) <= tols.h;
  const sMatch = Math.abs(c1.s - c2.s) <= tols.s;
  const lMatch = Math.abs(c1.l - c2.l) <= tols.l;
  return hMatch && sMatch && lMatch;
};

