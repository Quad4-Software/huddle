const PALETTES: [string, string][] = [
  ['#4f7df3', '#7b5cf0'],
  ['#3ecf8e', '#2bb3c0'],
  ['#f5a524', '#f56f4b'],
  ['#e5534b', '#d8417f'],
  ['#9b7bf0', '#5c7cf0'],
  ['#2bb3c0', '#3ecf8e'],
  ['#f56f9e', '#f5a524'],
  ['#5cc8f0', '#4f7df3'],
];

function hash(input: string): number {
  let h = 2166136261;
  for (let i = 0; i < input.length; i++) {
    h ^= input.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

export function avatarColors(name: string): [string, string] {
  const seed = hash(name || 'guest');
  return PALETTES[seed % PALETTES.length];
}

export function avatarInitials(name: string): string {
  const trimmed = (name || '').trim();
  if (!trimmed) return '?';
  const parts = trimmed.split(/\s+/).filter(Boolean);
  if (parts.length === 1) {
    return parts[0].slice(0, 2).toUpperCase();
  }
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function avatarGradient(name: string): string {
  const [a, b] = avatarColors(name);
  return `linear-gradient(135deg, ${a}, ${b})`;
}
