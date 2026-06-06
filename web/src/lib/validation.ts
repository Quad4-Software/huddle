export const MAX_ROOM_NAME_LENGTH = 80;
export const MAX_DISPLAY_NAME_LENGTH = 40;
export const MAX_PASSWORD_LENGTH = 256;
export const MAX_CHAT_MESSAGE_LENGTH = 4000;
export const MAX_FILE_SIZE = 25 * 1024 * 1024;

export function cleanBounded(value: string, max: number): string {
  return value.trim().slice(0, max);
}

export function fitsBound(value: string, max: number): boolean {
  return value.length <= max;
}
