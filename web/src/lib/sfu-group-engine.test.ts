/// <reference types="bun-types" />

import { describe, expect, test } from 'bun:test';

import { addVideoTransceiver, disableVideoTrack, diffVideoSubscriptions, groupVideoConstraints, mediaState, mergeRemoteStream, parseActiveSpeakers, preferredVideoCodecs, replaceEndedVideoTrack } from './sfu-group-engine';

class FakeTrack {
  enabled = true;
  readyState: MediaStreamTrackState = 'live';
  stopped = false;
  constructor(public id: string, public kind: 'audio' | 'video') {}
  stop() { this.stopped = true; this.readyState = 'ended'; }
}

class FakeStream {
  private tracks: FakeTrack[];

  constructor(tracks: FakeTrack[] = []) {
    this.tracks = [...tracks];
  }

  getTracks() { return [...this.tracks]; }
  getAudioTracks() { return this.tracks.filter(track => track.kind === 'audio'); }
  getVideoTracks() { return this.tracks.filter(track => track.kind === 'video'); }
  addTrack(track: FakeTrack) { this.tracks.push(track); }
  removeTrack(track: FakeTrack) { this.tracks = this.tracks.filter(candidate => candidate !== track); }
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

describe('addVideoTransceiver', () => {
  test('keeps the shared SFU connection bidirectional with one capped video encoding', () => {
    let init: RTCRtpTransceiverInit | undefined;
    const transceiver = { sender: {} } as RTCRtpTransceiver;
    const pc = {
      addTransceiver: (_track: MediaStreamTrack, value: RTCRtpTransceiverInit) => {
        init = value;
        return transceiver;
      },
    } as unknown as RTCPeerConnection;

    expect(addVideoTransceiver(pc, {} as MediaStreamTrack, {} as MediaStream)).toBe(transceiver);
    expect(init?.direction).toBe('sendrecv');
    expect(init?.sendEncodings).toEqual([{ maxBitrate: 700_000, maxFramerate: 24 }]);
    expect(init?.sendEncodings?.[0]?.rid).toBeUndefined();
  });

  test('falls back to a single video track when simulcast is rejected', () => {
    const fallback = { sender: {} } as RTCRtpTransceiver;
    const pc = {
      addTransceiver: () => { throw new TypeError('sendEncodings unsupported'); },
      addTrack: () => fallback.sender,
      getTransceivers: () => [fallback],
    } as unknown as RTCPeerConnection;

    expect(addVideoTransceiver(pc, {} as MediaStreamTrack, {} as MediaStream)).toBe(fallback);
  });
});

describe('groupVideoConstraints', () => {
  test('caps gallery capture before a fifth participant adds a fourth decoder', () => {
    expect(groupVideoConstraints()).toEqual({
      width: { ideal: 640, max: 640 },
      height: { ideal: 360, max: 360 },
      frameRate: { ideal: 24, max: 24 },
    });
  });
});

describe('camera lifecycle', () => {
  test('camera off detaches and stops capture without replacing the transceiver', async () => {
    const video = new FakeTrack('video-1', 'video');
    const stream = new FakeStream([video]);
    let replacedWith: MediaStreamTrack | null | undefined;
    const sender = { replaceTrack: async (track: MediaStreamTrack | null) => { replacedWith = track; } } as RTCRtpSender;

    await disableVideoTrack(stream as unknown as MediaStream, sender);

    expect(replacedWith).toBeNull();
    expect(video.stopped).toBeTrue();
    expect(stream.getVideoTracks()).toEqual([]);
  });

  test('does not advertise an enabled but ended camera track', () => {
    const video = new FakeTrack('video-1', 'video');
    video.readyState = 'ended';
    const stream = new FakeStream([video]);

    expect(mediaState(stream as unknown as MediaStream)).toEqual({ audio: false, video: false });
  });

  test('replaces an ended camera track on the existing sender', async () => {
    const ended = new FakeTrack('video-old', 'video');
    ended.readyState = 'ended';
    const replacement = new FakeTrack('video-new', 'video');
    const stream = new FakeStream([ended]);
    const senderState: { track: MediaStreamTrack | null } = { track: null };
    const sender = { replaceTrack: async (track: MediaStreamTrack | null) => { senderState.track = track; } } as RTCRtpSender;

    const track = await replaceEndedVideoTrack(stream as unknown as MediaStream, sender, async () => replacement as unknown as MediaStreamTrack);

    expect(track.id).toBe(replacement.id);
    expect(senderState.track?.id).toBe(replacement.id);
    expect(stream.getVideoTracks()).toEqual([replacement]);
  });
});
