import { SpotifyAdapter } from './SpotifyAdapter';
import type { PlayerAdapter, PlaybackStatus, Track, PlayerEvents } from './types';

class UnifiedPlayerController {
    currentTrack = $state<Track | null>(null);
    status = $state<PlaybackStatus>('idle');
    positionMs = $state(0);
    durationMs = $state(0);
    volume = $state(0.5);
    error = $state<string | null>(null);

    private adapters: Record<string, PlayerAdapter> = {};
    private activeAdapter: PlayerAdapter | null = null;
    private initialized = false;
    private initPromise: Promise<void> | null = null;

    constructor() {
        this.adapters = {
            Spotify: new SpotifyAdapter(),
        };
    }

    async init(onTrackEndedCallback: (id: number) => void) {
        if (this.initialized) return;
        if (this.initPromise) return this.initPromise;

        const events: PlayerEvents = {
            onTrackEnded: (roomTrackId) => {
                this.status = 'idle';
                this.currentTrack = null;
                onTrackEndedCallback(roomTrackId);
            },
            onStatusChange: (status) => {
                this.status = status;
            },
            onProgress: (pos, dur) => {
                this.positionMs = pos;
                this.durationMs = dur;
            },
            onError: (err) => {
                this.error = err;
            }
        };

        this.initPromise = Promise.all(
            Object.values(this.adapters).map(a => a.initialize(events))
        ).then(() => {
            this.initialized = true;
        }).catch(err => {
            this.error = "Initialization failed: " + err;
        });

        return this.initPromise;
    }

    async play(track: Track) {
        if (!this.initialized) {
            if (this.initPromise) await this.initPromise;
            else throw new Error('Player not initialized');
        }

        const nextAdapter = this.adapters[track.platform];
        if (!nextAdapter) {
            this.error = `Unsupported platform: ${track.platform}`;
            return;
        }

        if (this.activeAdapter && this.activeAdapter !== nextAdapter) {
            await this.activeAdapter.pause();
        }

        this.activeAdapter = nextAdapter;
        this.currentTrack = track;
        
        try {
            await this.activeAdapter.play(track);
            await this.activeAdapter.setVolume(this.volume);
        } catch (err: any) {
            this.error = err.message || 'Playback failed';
            this.status = 'idle';
        }
    }

    async pause() {
        if (this.activeAdapter && this.status === 'playing') {
            await this.activeAdapter.pause();
            this.status = 'paused';
        }
    }

    async resume() {
        if (this.activeAdapter && this.status === 'paused') {
            await this.activeAdapter.resume();
            this.status = 'playing';
        }
    }
    
    async togglePlay() {
        if (this.status === 'playing') {
            await this.pause();
        } else if (this.status === 'paused') {
            await this.resume();
        }
    }

    async seek(ms: number) {
        if (this.activeAdapter) {
            await this.activeAdapter.seek(ms);
            this.positionMs = ms;
        }
    }

    async setVolume(vol: number) {
        this.volume = vol;
        if (this.activeAdapter) {
            await this.activeAdapter.setVolume(vol);
        }
    }

    async stop() {
        if (this.activeAdapter) {
            await this.activeAdapter.pause();
        }
        this.status = 'idle';
        this.currentTrack = null;
    }

    async disconnect() {
        await Promise.all(
            Object.values(this.adapters).map(a => a.disconnect())
        );
    }
}

export const globalPlayer = new UnifiedPlayerController();
