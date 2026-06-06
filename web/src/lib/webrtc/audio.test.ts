import { describe, expect, it } from 'vitest';
import { isSpeaking, measureAudioLevel } from './audio';

function mockAnalyser(values: number[]): AnalyserNode {
  return {
    frequencyBinCount: values.length,
    getByteFrequencyData: (out: Uint8Array) => {
      out.set(values);
    },
  } as AnalyserNode;
}

describe('isSpeaking', () => {
  it('detects energy above threshold', () => {
    expect(isSpeaking(mockAnalyser([40, 40, 40, 40]), 15)).toBe(true);
  });

  it('stays silent below threshold', () => {
    expect(isSpeaking(mockAnalyser([0, 1, 0, 1]), 15)).toBe(false);
  });

  it('uses custom threshold', () => {
    const analyser = mockAnalyser([10, 10, 10, 10]);
    expect(isSpeaking(analyser, 5)).toBe(true);
    expect(isSpeaking(analyser, 20)).toBe(false);
  });
});

describe('measureAudioLevel', () => {
  it('returns normalized level from analyser data', () => {
    expect(measureAudioLevel(mockAnalyser([0, 0, 0, 0]))).toBe(0);
    expect(measureAudioLevel(mockAnalyser([72, 72, 72, 72]))).toBe(1);
    expect(measureAudioLevel(mockAnalyser([36, 36, 36, 36]))).toBeCloseTo(0.5, 1);
  });
});
