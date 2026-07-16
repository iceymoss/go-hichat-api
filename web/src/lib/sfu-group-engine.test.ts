/// <reference types="bun-types" />

import { describe, expect, test } from 'bun:test';

import { diffVideoSubscriptions, mergeRemoteStream, parseActiveSpeakers, preferredVideoCodecs, videoSendEncodings } from './sfu-group-engine';

class FakeTrack {
  constructor(public id: string, public kind: 'audio' | 'video') {}
}

class FakeStream {
  private tracks: FakeTrack[];

  constructor(tracks: FakeTrack[] = []) {
    this.tracks = [...tracks];
  }

  getTracks() { return [...this.tracks]; }
  addTrack(track: FakeTrack) { this.tracks.push(track); }
}

describe('mergeRemoteStream', () => {
  test.each([
    ['video then audio', new FakeTrack('video-1', 'video'), new FakeTrack('audio-1', 'audio')],
    ['audio then video', new FakeTrack('audio-1', 'audio'), new FakeTrack('video-1', 'video')],
  ])('%s keeps both tracks', (_name, first, second) => {
    const aggregate = new FakeStream();
    mergeRemoteStream(aggregate as unknown as MediaStream, new FakeStream([first]) as unknown as MediaStream, first as unknown as MediaStreamTrack);
    mergeRemoteStream(aggregate as unknown as MediaStream, new FakeStream([second]) as unknown as MediaStream, second as unknown as MediaStreamTrack);

    expect(aggregate.getTracks().map(track => track.kind).sort()).toEqual(['audio', 'video']);
  });

  test('does not duplicate a track present in both event and stream', () => {
    const video = new FakeTrack('video-1', 'video');
    const aggregate = new FakeStream();
    mergeRemoteStream(aggregate as unknown as MediaStream, new FakeStream([video]) as unknown as MediaStream, video as unknown as MediaStreamTrack);

    expect(aggregate.getTracks()).toHaveLength(1);
  });
});

describe('preferredVideoCodecs', () => {
  test('forces VP8 when browsers also advertise hardware H264', () => {
    const codecs = [
      { mimeType: 'video/H264', clockRate: 90000 },
      { mimeType: 'video/VP8', clockRate: 90000 },
      { mimeType: 'video/VP9', clockRate: 90000 },
    ] as RTCRtpCodec[];

    expect(preferredVideoCodecs(codecs).map(codec => codec.mimeType)).toEqual(['video/VP8']);
  });

  test('returns no preference when VP8 is unavailable', () => {
    const codecs = [{ mimeType: 'video/H264', clockRate: 90000 }] as RTCRtpCodec[];

    expect(preferredVideoCodecs(codecs)).toEqual([]);
  });
});

describe('parseActiveSpeakers', () => {
  test.each([
    [{ speakers: ['12', '11', '12', '', 20] }, ['12', '11']],
    [{ speakers: [] }, []],
    [{ speakers: '12' }, []],
    [{}, []],
  ])('sanitizes %o', (data, want) => {
    expect(parseActiveSpeakers(data)).toEqual(want);
  });
});

describe('diffVideoSubscriptions', () => {
  test('emits only page changes', () => {
    expect(diffVideoSubscriptions(new Set(['11', '12']), new Set(['12', '20']))).toEqual({
      subscribe: ['20'],
      unsubscribe: ['11'],
    });
  });

  test('does nothing for an unchanged page', () => {
    expect(diffVideoSubscriptions(new Set(['11']), new Set(['11']))).toEqual({ subscribe: [], unsubscribe: [] });
  });
});

describe('videoSendEncodings', () => {
  test('publishes q h f VP8 simulcast layers', () => {
    expect(videoSendEncodings()).toEqual([
      { rid: 'q', scaleResolutionDownBy: 4, maxBitrate: 150000 },
      { rid: 'h', scaleResolutionDownBy: 2, maxBitrate: 500000 },
      { rid: 'f', scaleResolutionDownBy: 1, maxBitrate: 1500000 },
    ]);
  });
});
