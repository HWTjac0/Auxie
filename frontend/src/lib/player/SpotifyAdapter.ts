import type { PlayerAdapter, PlayerEvents, Track } from './types';

export class SpotifyAdapter implements PlayerAdapter {
    private player: any = null;
    private deviceId: string | null = null;
    private token: string | null = null;
    private events!: PlayerEvents;
    private trackCheckInterval: number | null = null;
    private currentTrack: Track | null = null;

    async initialize(events: PlayerEvents): Promise<void> {
        this.events = events;
        
        try {
            const meRes = await fetch("/api/v1/auth/me");
            if (meRes.ok) {
                const me = await meRes.json();
                if (!me.spotify_name) {
                    console.log("ℹ️ Spotify not connected for this user, skipping Spotify SDK initialization");
                    return;
                }
            } else {
                return;
            }
        } catch (e) {
            console.error("❌ Failed to fetch user auth status", e);
            return;
        }

        return new Promise((resolve, reject) => {
            (window as any).onSpotifyWebPlaybackSDKReady = () => {
                this.initSpotifyPlayer().then(resolve).catch(reject);
            };

            if (!(window as any).Spotify) {
                const script = document.createElement("script");
                script.src = "https://sdk.scdn.co/spotify-player.js";
                script.async = true;
                document.head.appendChild(script);
            } else {
                this.initSpotifyPlayer().then(resolve).catch(reject);
            }
        });
    }

    private async initSpotifyPlayer(): Promise<void> {
        return new Promise(async (resolve, reject) => {
            try {
                const tokenRes = await fetch("/api/v1/playback/token");
                if (!tokenRes.ok) {
                    const msg = "Could not fetch Spotify token";
                    this.events.onError(msg);
                    return reject(new Error(msg));
                }

                const { access_token } = await tokenRes.json();
                this.token = access_token;

                if (!(window as any).Spotify) {
                    const msg = "Spotify SDK not ready";
                    this.events.onError(msg);
                    return reject(new Error(msg));
                }

                this.player = new (window as any).Spotify.Player({
                    name: "Auxie",
                    getOAuthToken: (cb: (token: string) => void) => {
                        cb(access_token);
                    },
                    volume: 0.5,
                });

                this.player.addListener("initialization_error", ({ message }: { message: string }) => {
                    this.events.onError(message);
                    reject(new Error(message));
                });
                this.player.addListener("authentication_error", ({ message }: { message: string }) => {
                    this.events.onError("Authentication error: " + message);
                    reject(new Error("Authentication error: " + message));
                });
                this.player.addListener("account_error", ({ message }: { message: string }) => {
                    this.events.onError("Account error: " + message);
                });
                this.player.addListener("playback_error", ({ message }: { message: string }) => {
                    this.events.onError("Playback error: " + message);
                });

                this.player.addListener("player_state_changed", (state: any) => {
                    if (state) {
                        const currentTrack = state.track_window?.current_track;
                        if (currentTrack && !state.paused) {
                            this.events.onStatusChange("playing");
                        } else if (state.paused) {
                            this.events.onStatusChange("paused");
                        }

                        this.events.onProgress(state.position, state.duration);

                        if (state.position === 0 && state.duration > 0 && state.paused && this.currentTrack) {
                            this.events.onTrackEnded(this.currentTrack.room_track_id);
                            this.currentTrack = null;
                        }
                    }
                });

                this.player.addListener("ready", ({ device_id }: { device_id: string }) => {
                    this.deviceId = device_id;
                    resolve();
                });

                this.player.addListener("not_ready", ({ device_id }: { device_id: string }) => {
                    if (this.deviceId === device_id) this.deviceId = null;
                });

                await this.player.connect();
            } catch (err: any) {
                this.events.onError("Failed to initialize Spotify player: " + err.message);
                reject(err);
            }
        });
    }

    async play(track: Track): Promise<void> {
        this.currentTrack = track;
        
        if (!this.player || !this.deviceId || !this.token) {
            throw new Error("Spotify player not fully initialized");
        }

        const playUrl = `https://api.spotify.com/v1/me/player/play?device_id=${this.deviceId}`;
        const response = await fetch(playUrl, {
            method: "PUT",
            headers: {
                "Authorization": `Bearer ${this.token}`,
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ uris: [track.source_uri] }),
        });

        if (!response.ok) {
            if (response.status === 404) {
                // Try transferring playback
                const transferRes = await fetch("https://api.spotify.com/v1/me/player", {
                    method: "PUT",
                    headers: {
                        "Authorization": `Bearer ${this.token}`,
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({ device_ids: [this.deviceId], play: true }),
                });

                if (transferRes.ok) {
                    const retry = await fetch(playUrl, {
                        method: "PUT",
                        headers: {
                            "Authorization": `Bearer ${this.token}`,
                            "Content-Type": "application/json",
                        },
                        body: JSON.stringify({ uris: [track.source_uri] }),
                    });
                    
                    if (!retry.ok) {
                        throw new Error("Failed to retry play after transfer");
                    }
                } else {
                    throw new Error("Failed to transfer playback");
                }
            } else {
                throw new Error(`Spotify API error: ${response.status}`);
            }
        }

        this.events.onStatusChange("playing");

        // Polling interval for position and end of track
        if (this.trackCheckInterval !== null) {
            window.clearInterval(this.trackCheckInterval);
        }

        this.trackCheckInterval = window.setInterval(async () => {
            if (!this.player) return;
            const state = await this.player.getCurrentState();
            if (state === null) {
                if (this.trackCheckInterval !== null) {
                    window.clearInterval(this.trackCheckInterval);
                    this.trackCheckInterval = null;
                }
                return;
            }

            this.events.onProgress(state.position, state.duration);

            if (state.position >= state.duration - 1000 && !state.paused) {
                if (this.trackCheckInterval !== null) {
                    window.clearInterval(this.trackCheckInterval);
                    this.trackCheckInterval = null;
                }
                if (this.currentTrack) {
                    this.events.onTrackEnded(this.currentTrack.room_track_id);
                    this.currentTrack = null;
                }
            }
        }, 500);
    }

    async pause(): Promise<void> {
        if (this.player) {
            await this.player.pause();
        }
    }

    async resume(): Promise<void> {
        if (this.player) {
            await this.player.resume();
        }
    }

    async seek(positionMs: number): Promise<void> {
        if (this.player) {
            await this.player.seek(positionMs);
        }
    }

    async setVolume(volume: number): Promise<void> {
        if (this.player) {
            await this.player.setVolume(volume);
        }
    }

    async disconnect(): Promise<void> {
        if (this.trackCheckInterval !== null) {
            window.clearInterval(this.trackCheckInterval);
            this.trackCheckInterval = null;
        }
        if (this.player) {
            this.player.disconnect();
            this.player = null;
        }
    }
}
