// WCAG relative-luminance contrast ratio between two colors, each given as
// "#rrggbb" or "rgb(r, g, b)". Pure, dependency-free, single responsibility
// (see .context/adr/20260712-233457-theme-contrast-hierarchy-fix.md).
export function contrastRatio(colorA, colorB) {
  const luminanceA = relativeLuminance(parseColor(colorA));
  const luminanceB = relativeLuminance(parseColor(colorB));
  const lighter = Math.max(luminanceA, luminanceB);
  const darker = Math.min(luminanceA, luminanceB);
  return (lighter + 0.05) / (darker + 0.05);
}

function parseColor(color) {
  const rgbMatch = color.match(/rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)/i);
  if (rgbMatch) {
    return [Number(rgbMatch[1]), Number(rgbMatch[2]), Number(rgbMatch[3])];
  }
  const longHexMatch = color.match(/^#?([0-9a-f]{6})$/i);
  if (longHexMatch) {
    const hex = longHexMatch[1];
    return [parseInt(hex.slice(0, 2), 16), parseInt(hex.slice(2, 4), 16), parseInt(hex.slice(4, 6), 16)];
  }
  const shortHexMatch = color.match(/^#?([0-9a-f]{3})$/i);
  if (shortHexMatch) {
    const [r, g, b] = shortHexMatch[1];
    return [parseInt(r + r, 16), parseInt(g + g, 16), parseInt(b + b, 16)];
  }
  throw new Error(`contrastRatio: unrecognized color format "${color}"`);
}

function relativeLuminance([r, g, b]) {
  const channel = (value) => {
    const normalized = value / 255;
    return normalized <= 0.03928 ? normalized / 12.92 : Math.pow((normalized + 0.055) / 1.055, 2.4);
  };
  const [rl, gl, bl] = [channel(r), channel(g), channel(b)];
  return 0.2126 * rl + 0.7152 * gl + 0.0722 * bl;
}
