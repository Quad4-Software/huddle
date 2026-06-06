export interface PowChallenge {
  id: string;
  prefix: string;
  difficulty: number;
  expiresAt: number;
}

export interface PowSolution {
  id: string;
  nonce: number;
}

export async function fetchChallenge(action: 'create' | 'join'): Promise<PowChallenge | null> {
  const res = await fetch(`/api/pow/challenge?action=${action}`);
  if (res.status === 204) return null;
  if (!res.ok) {
    throw new Error(`Could not fetch proof-of-work challenge (${res.status})`);
  }
  return (await res.json()) as PowChallenge;
}

function hasLeadingZeroBits(hash: Uint8Array, bits: number): boolean {
  if (bits <= 0) return true;
  const fullBytes = Math.floor(bits / 8);
  const remBits = bits % 8;
  for (let i = 0; i < fullBytes; i++) {
    if (hash[i] !== 0) return false;
  }
  if (remBits === 0) return true;
  const mask = 0xff << (8 - remBits);
  return (hash[fullBytes] & mask) === 0;
}

export async function solveChallenge(
  prefix: string,
  difficulty: number,
  onProgress?: (progress: number) => void,
): Promise<number> {
  const expected = Math.max(1, 2 ** difficulty);
  let nonce = 0;
  while (true) {
    const data = new TextEncoder().encode(`${prefix}:${nonce}`);
    const digest = await crypto.subtle.digest('SHA-256', data);
    if (hasLeadingZeroBits(new Uint8Array(digest), difficulty)) {
      onProgress?.(100);
      return nonce;
    }
    nonce++;
    if (nonce % 4096 === 0) {
      onProgress?.(Math.min(95, (nonce / expected) * 100));
      await new Promise((resolve) => setTimeout(resolve, 0));
    }
  }
}

export async function solvePow(
  action: 'create' | 'join',
  onProgress?: (progress: number) => void,
): Promise<PowSolution | undefined> {
  const challenge = await fetchChallenge(action);
  if (!challenge) return undefined;
  const nonce = await solveChallenge(challenge.prefix, challenge.difficulty, onProgress);
  return { id: challenge.id, nonce };
}
