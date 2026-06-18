import type { PlayerAdapter, PlayerEvents, Track } from './types';

export class TidalAdapter implements PlayerAdapter {
    private audio: HTMLAudioElement | null = null;
    private events!: PlayerEvents;
    private currentTrack: Track | null = null;
    private slug: string | null = null;

    async initialize(events: PlayerEvents): Promise<void> {
        this.events = events;
        
        // Extract room slug from URL to construct stream URL
        const pathParts = window.location.pathname.split('/');
        if (pathParts.length >= 3 && pathParts[1] === 'room') {
            this.slug = pathParts[2];
        }

        this.audio = new Audio();
        
        this.audio.addEventListener('play', () => {
            this.events.onStatusChange('playing');
        });

        this.audio.addEventListener('pause', () => {
            this.events.onStatusChange('paused');
        });

        this.audio.addEventListener('timeupdate', () => {
            if (this.audio) {
                // Audio position is in seconds, convert to ms
                this.events.onProgress(this.audio.currentTime * 1000, this.audio.duration * 1000);
            }
        });

        this.audio.addEventListener('ended', () => {
            if (this.currentTrack) {
                this.events.onTrackEnded(this.currentTrack.room_track_id);
                this.currentTrack = null;
            }
        });

        this.audio.addEventListener('error', (e) => {
            console.error("Tidal Audio Error", e);
            this.events.onError("Failed to play Tidal stream. Backend streaming may not be fully implemented.");
            this.events.onStatusChange('idle');
        });
        
        return Promise.resolve();
    }

    async play(track: Track): Promise<void> {
        if (!this.audio) throw new Error("Tidal adapter not initialized");
        if (!this.slug) throw new Error("Room slug not found");

        this.currentTrack = track;
        
        // Construct the backend stream URL for Tidal
        this.audio.src = `/api/v1/stream/tidal/${track.room_track_id}`;
        
        try {
            await this.audio.play();
        } catch (err: any) {
            this.events.onError("Playback error: " + err.message);
        }
    }

    async pause(): Promise<void> {
        if (this.audio) {
            this.audio.pause();
        }
    }

    async resume(): Promise<void> {
        if (this.audio) {
            await this.audio.play();
        }
    }

    async seek(positionMs: number): Promise<void> {
        if (this.audio) {
            this.audio.currentTime = positionMs / 1000;
        }
    }

    async setVolume(volume: number): Promise<void> {
        if (this.audio) {
            // HTMLAudioElement volume is between 0.0 and 1.0
            this.audio.volume = Math.max(0, Math.min(1, volume));
        }
    }

    async disconnect(): Promise<void> {
        if (this.audio) {
            this.audio.pause();
            this.audio.src = '';
            this.audio = null;
        }
    }
}
