const ICE_SERVERS: RTCIceServer[] = [{ urls: 'stun:stun.l.google.com:19302' }];

export function createPeer(): RTCPeerConnection {
  const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS });
  preferOpus(pc);
  return pc;
}

function preferOpus(pc: RTCPeerConnection) {
  const apply = () => {
    try {
      for (const t of pc.getTransceivers()) {
        if (t.sender.track?.kind === 'audio') {
          const caps = RTCRtpSender.getCapabilities('audio');
          const opus = caps?.codecs.find((c) => c.mimeType.toLowerCase() === 'audio/opus');
          if (opus && caps) {
            t.setCodecPreferences([opus, ...caps.codecs.filter((c) => c !== opus)]);
          }
        }
      }
    } catch {}
  };
  pc.addEventListener('track', apply);
  pc.addEventListener('negotiationneeded', apply);
}
