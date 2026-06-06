export async function generateInviteQrDataUrl(url: string): Promise<string> {
  const QRCode = await import('qrcode');
  return QRCode.toDataURL(url, {
    errorCorrectionLevel: 'M',
    margin: 2,
    width: 280,
    color: {
      dark: '#e8eaedff',
      light: '#161b22ff',
    },
  });
}
