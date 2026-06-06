const encoder = new TextEncoder();
const decoder = new TextDecoder();

function decodeBase64Url(raw: string): Uint8Array {
  let b64 = raw.replace(/-/g, '+').replace(/_/g, '/');
  while (b64.length % 4) b64 += '=';
  return Uint8Array.from(atob(b64), (c) => c.charCodeAt(0));
}

function encodeBase64Url(bytes: Uint8Array): string {
  return btoa(String.fromCharCode(...bytes))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

export async function importRoomKey(raw: string): Promise<CryptoKey> {
  const bytes = decodeBase64Url(raw);
  return crypto.subtle.importKey('raw', bytes, { name: 'AES-GCM' }, false, ['encrypt', 'decrypt']);
}

export async function importSigningKey(raw: string): Promise<CryptoKey> {
  const bytes = decodeBase64Url(raw);
  return crypto.subtle.importKey('raw', bytes, { name: 'HMAC', hash: 'SHA-256' }, false, [
    'sign',
    'verify',
  ]);
}

function encodeBase64(bytes: Uint8Array): string {
  let binary = '';
  for (let i = 0; i < bytes.length; i += 0x8000) {
    binary += String.fromCharCode(...bytes.subarray(i, i + 0x8000));
  }
  return btoa(binary);
}

export async function encrypt(key: CryptoKey, data: Uint8Array | string): Promise<string> {
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const plain = typeof data === 'string' ? encoder.encode(data) : data;
  const cipher = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, plain);
  const out = new Uint8Array(iv.length + cipher.byteLength);
  out.set(iv);
  out.set(new Uint8Array(cipher), iv.length);
  return encodeBase64(out);
}

export async function decrypt(key: CryptoKey, encoded: string): Promise<Uint8Array> {
  const raw = Uint8Array.from(atob(encoded), (c) => c.charCodeAt(0));
  const iv = raw.slice(0, 12);
  const cipher = raw.slice(12);
  const plain = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, cipher);
  return new Uint8Array(plain);
}

export async function decryptText(key: CryptoKey, encoded: string): Promise<string> {
  const bytes = await decrypt(key, encoded);
  return decoder.decode(bytes);
}

export async function signText(key: CryptoKey, text: string): Promise<string> {
  const sig = await crypto.subtle.sign('HMAC', key, encoder.encode(text));
  return encodeBase64Url(new Uint8Array(sig));
}

export async function verifyText(
  key: CryptoKey,
  text: string,
  signature: string,
): Promise<boolean> {
  return crypto.subtle.verify('HMAC', key, decodeBase64Url(signature), encoder.encode(text));
}
