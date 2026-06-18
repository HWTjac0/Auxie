export interface Track {
    room_track_id: number;
    title: string;
    artist: string;
    source_uri: string;
    platform: 'Spotify' | 'Tidal' | 'SoundCloud';
    duration_ms?: number;
    cover_url?: string;
    like_count?: number;
}

export type PlaybackStatus = 'idle' | 'playing' | 'paused' | 'loading';

export interface PlayerEvents {
    onTrackEnded: (roomTrackId: number) => void;
    onStatusChange: (status: PlaybackStatus) => void;
    onProgress: (positionMs: number, durationMs: number) => void;
    onError: (error: string) => void;
}

export interface PlayerAdapter {
    initialize(events: PlayerEvents): Promise<void>;
    play(track: Track): Promise<void>;
    pause(): Promise<void>;
    resume(): Promise<void>;
    seek(positionMs: number): Promise<void>;
    setVolume(volume: number): Promise<void>;
    disconnect(): Promise<void>;
}
