<script lang="ts">
import { onMount } from "svelte";
import SkipForward from "./icons/SkipForward.svelte";

interface PlaybackState {
	track: any;
	status: 'idle' | 'playing' | 'paused';
	startedAt: string | null;
}

interface SpotifyPlayer {
	addListener: (listener: any) => void;
	getCurrentState: () => Promise<any>;
	play: (options: any) => Promise<void>;
	pause: () => Promise<void>;
	disconnect: () => void;
}

declare global {
	interface Window {
		onSpotifyWebPlaybackSDKReady?: () => void;
		Spotify?: {
			Player: new (options: any) => SpotifyPlayer;
		};
	}
}

let { queue = [], currentUser, slug, ws }: { queue?: any[], currentUser?: any, slug?: string, ws?: WebSocket } = $props();

let playback: PlaybackState = $state({
	track: null,
	status: 'idle',
	startedAt: null,
});

let spotifyPlayer: SpotifyPlayer | null = $state(null);
let spotifyDeviceID: string | null = $state(null);
let spotifyToken: string | null = $state(null);
let isSkipping = $state(false);
let trackCheckInterval: number | null = $state(null);
let canManage = $derived(currentUser?.CurrentRole === "Host" || currentUser?.CurrentRole === "DJ");

// Load Spotify Web Playback SDK
function loadSpotifySDK() {
	if (window.Spotify) {
		console.log("✅ Spotify SDK already loaded");
		return true;
	}

	console.log("📥 Loading Spotify Web Playback SDK...");
	const script = document.createElement("script");
	script.src = "https://sdk.scdn.co/spotify-player.js";
	script.async = true;
	document.head.appendChild(script);
	return false;
}

// Initialize Spotify Player
async function initSpotifyPlayer() {
	try {
		console.log("🔄 Initializing Spotify Player...");
		const tokenRes = await fetch("/api/v1/playback/token");
		if (!tokenRes.ok) {
			console.error("❌ Could not fetch Spotify token", tokenRes.statusText);
			return;
		}

		const { access_token } = await tokenRes.json();
		spotifyToken = access_token;
		console.log("✅ Spotify token received", access_token.substring(0, 10) + "...");

		if (!window.Spotify) {
			console.error("❌ Spotify SDK not ready");
			return;
		}

		console.log("🎵 Creating Spotify Player instance...");
		spotifyPlayer = new window.Spotify.Player({
			name: "Auxie",
			getOAuthToken: (cb: (token: string) => void) => {
				console.log("🔑 Spotify requesting token callback");
				cb(access_token);
			},
			volume: 0.5,
		});

		// Errors
		spotifyPlayer.addListener("initialization_error", ({ message }: { message: string }) => {
			console.error("Initialization Error", message);
		});
		spotifyPlayer.addListener("authentication_error", ({ message }: { message: string }) => {
			console.error("Authentication Error", message);
		});
		spotifyPlayer.addListener("account_error", ({ message }: { message: string }) => {
			console.error("Account Error", message);
		});
		spotifyPlayer.addListener("playback_error", ({ message }: { message: string }) => {
			console.error("Playback Error", message);
		});

		// Playback status updates
		spotifyPlayer.addListener("player_state_changed", (state: any) => {
			if (state) {
				const currentTrack = state.track_window?.current_track;
				if (currentTrack && !state.paused) {
					playback.status = "playing";
				} else if (state.paused) {
					playback.status = "paused";
				}

				// Check if track ended
				if (state.position === 0 && state.duration > 0 && playback.status === "paused" && playback.track) {
					console.log("Track ended via SDK");
					notifyPlaybackEnded(playback.track.room_track_id);
					playback.status = "idle";
				}
			}
		});

		// Connect player
		console.log("📡 Connecting Spotify Player...");
		const connectPromise = spotifyPlayer.connect();

		// Spotify Web Playback SDK emits a 'ready' event with the device id.
		// Attach listeners to capture the device id once ready.
		spotifyPlayer.addListener("ready", ({ device_id }: { device_id: string }) => {
			spotifyDeviceID = device_id;
			console.log("✅ Spotify Player ready — Device ID:", spotifyDeviceID);
		});

		spotifyPlayer.addListener("not_ready", ({ device_id }: { device_id: string }) => {
			console.log("🔌 Spotify device went offline:", device_id);
			if (spotifyDeviceID === device_id) spotifyDeviceID = null;
		});

		if (connectPromise && typeof connectPromise.then === 'function') {
			connectPromise.then((success: boolean) => {
				if (success) {
					console.log("✅ Spotify Player connected");
				} else {
					console.error("❌ Failed to connect Spotify Player");
				}
			}).catch((err: any) => {
				console.error("❌ Error connecting Spotify Player:", err);
			});
		} else {
			console.error("❌ connect() did not return a promise");
		}
	} catch (err) {
		console.error("❌ Failed to initialize Spotify player:", err);
	}
}

// Listen to WS events
function setupWSListener() {
	if (!ws) return;

	const originalOnMessage = ws.onmessage;

	ws.onmessage = (event) => {
		try {
			const msg = JSON.parse(event.data);

			if (msg.type === "playback:start") {
				playback.track = msg.track;
				playback.status = "playing";
				playback.startedAt = msg.started_at;

				console.log("🎵 Playback started:", msg.track.title, "Platform:", msg.track.platform);
				startPlayingTrack(msg.track);
			} else if (msg.type === "playback:skipped") {
				stopPlayback();
				playback.status = "idle";
				console.log("⏭️  Track skipped");
			} else if (msg.type === "playback:ended") {
				stopPlayback();
				playback.status = "idle";
				console.log("✅ Track ended");
			}
		} catch (err) {
			console.error("Error parsing WS message:", err);
		}

		// Call original handler if exists
		if (originalOnMessage) {
			originalOnMessage.call(ws, event);
		}
	};
}

async function startPlayingTrack(track: any) {
	if (!track) return;

	console.log("▶️  Starting playback for:", track.title, "Platform:", track.platform);
	console.log("📊 Playback state:", {
		spotifyPlayer: !!spotifyPlayer,
		spotifyDeviceID,
		spotifyToken: spotifyToken ? "present" : "missing",
		track_uri: track.source_uri,
	});

	// Clear previous interval if any
	if (trackCheckInterval !== null) {
		clearInterval(trackCheckInterval);
		trackCheckInterval = null;
	}

	// For Spotify tracks - use Web Playback SDK
	if (track.platform === "Spotify" && spotifyPlayer && spotifyDeviceID && spotifyToken) {
		try {
			console.log("🎶 Playing Spotify track via Web Playback SDK:", track.source_uri);
			
			// Use Spotify API to play on specific device. The play endpoint accepts
			// a `device_id` query parameter; `device_ids` belongs to the transfer
			// endpoint. First try starting playback on the device, then fall back
			// to transferring playback if Spotify reports no active device.
			const playUrl = `https://api.spotify.com/v1/me/player/play?device_id=${spotifyDeviceID}`;
			const response = await fetch(playUrl, {
				method: "PUT",
				headers: {
					"Authorization": `Bearer ${spotifyToken}`,
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ uris: [track.source_uri] }),
			});

			if (!response.ok) {
				console.error("❌ Spotify API error:", response.status, response.statusText);
				const errBody = await response.text();
				console.error("Response:", errBody);

				// If there's no active device, transfer playback to our SDK device
				// and retry starting playback.
				if (response.status === 404) {
					try {
						const transferRes = await fetch("https://api.spotify.com/v1/me/player", {
							method: "PUT",
							headers: {
								"Authorization": `Bearer ${spotifyToken}`,
								"Content-Type": "application/json",
							},
							body: JSON.stringify({ device_ids: [spotifyDeviceID], play: true }),
						});

						if (transferRes.ok) {
							// Retry play after successful transfer
							const retry = await fetch(playUrl, {
								method: "PUT",
								headers: {
									"Authorization": `Bearer ${spotifyToken}`,
									"Content-Type": "application/json",
								},
								body: JSON.stringify({ uris: [track.source_uri] }),
							});

							if (!retry.ok) {
								console.error("❌ Spotify retry error:", retry.status, retry.statusText);
								const retryBody = await retry.text();
								console.error("Retry response:", retryBody);
								playback.status = "idle";
								return;
							}
						} else {
							console.error("❌ Failed to transfer playback:", transferRes.status, transferRes.statusText);
							const tBody = await transferRes.text();
							console.error("Transfer response:", tBody);
							playback.status = "idle";
							return;
						}
					} catch (err) {
						console.error("❌ Error transferring playback:", err);
						playback.status = "idle";
						return;
					}
				} else {
					playback.status = "idle";
					return;
				}
			}

			playback.status = "playing";
			console.log("✅ Track now playing via Spotify!");

			// Listen for track end
			trackCheckInterval = setInterval(async () => {
				if (!spotifyPlayer) return;
				
				const state = await spotifyPlayer.getCurrentState();
				if (state === null) {
					// Playback device has been disconnected
					if (trackCheckInterval !== null) {
						clearInterval(trackCheckInterval);
						trackCheckInterval = null;
					}
					return;
				}

				// Check if we're at the end of the track (within 1 second of duration)
				if (state.position >= state.duration - 1000 && !state.paused) {
					if (trackCheckInterval !== null) {
						clearInterval(trackCheckInterval);
						trackCheckInterval = null;
					}
					notifyPlaybackEnded(track.room_track_id);
					playback.status = "idle";
				}
			}, 500);
		} catch (err) {
			console.error("Error playing Spotify track:", err);
			playback.status = "idle";
		}
		return;
	}

	// For other platforms - show not implemented
	console.warn(`Playback for ${track.platform} not yet implemented`);
	playback.status = "idle";
}

function stopPlayback() {
	if (spotifyPlayer) {
		spotifyPlayer.pause().catch(console.error);
	}
}

function notifyPlaybackEnded(roomTrackId: number) {
	if (ws && ws.readyState === WebSocket.OPEN) {
		ws.send(
			JSON.stringify({
				type: "playback:ended",
				room_track_id: roomTrackId,
			})
		);
	}
}

async function skipTrack() {
	if (!slug || isSkipping || !canManage) return;
	isSkipping = true;
	try {
		const res = await fetch(`/api/v1/room/${slug}/skip`, { method: "POST" });
		if (!res.ok) console.error("Failed to skip track");
	} catch (err) {
		console.error(err);
	} finally {
		isSkipping = false;
	}
}

// Initialize on mount
onMount(() => {
	console.log("🚀 NowPlaying component mounted");
	
	// Set callback FIRST (before loading SDK)
	window.onSpotifyWebPlaybackSDKReady = () => {
		console.log("✅ Spotify SDK Ready callback triggered");
		initSpotifyPlayer();
	};

	// Then load SDK
	const alreadyLoaded = loadSpotifySDK();
	
	// If SDK is already loaded, trigger init immediately
	if (alreadyLoaded && window.Spotify) {
		console.log("⚡ SDK was cached, initializing immediately");
		initSpotifyPlayer();
	}
});

// Reactive trigger - set listener when ws changes
$effect(() => {
	if (ws) {
		setupWSListener();
	}
});

// Cleanup on unmount
$effect.pre(() => {
	return () => {
		stopPlayback();
		if (trackCheckInterval !== null) {
			clearInterval(trackCheckInterval);
			trackCheckInterval = null;
		}
		if (spotifyPlayer) {
			spotifyPlayer.disconnect();
		}
	};
});
</script>

<div class="now-playing">
	{#if playback.track}
		<h3 class="section-title">Now Playing</h3>
		<div class="playing-card">
			<img 
				src={playback.track.cover_url?.String || playback.track.cover_url || "/placeholder.png"} 
				alt={playback.track.title} 
				class="playing-cover"
			/>
			<div class="playing-info">
				<h4>{playback.track.title}</h4>
				<p>{playback.track.artist?.String || playback.track.artist || "Unknown Artist"}</p>
				<p class="status-text">
					{#if playback.status === 'playing'}
						<span class="status-badge playing">🎵 Playing</span>
					{:else if playback.status === 'paused'}
						<span class="status-badge paused">⏸️ Paused</span>
					{:else}
						<span class="status-badge idle">⏹️ Idle</span>
					{/if}
				</p>
			</div>
			<div class="playing-actions">
				<div class="playing-indicator">
					{#if playback.status === 'playing'}
						<div class="bar"></div>
						<div class="bar"></div>
						<div class="bar"></div>
					{:else}
						<div class="bar paused"></div>
						<div class="bar paused"></div>
						<div class="bar paused"></div>
					{/if}
				</div>
				
				{#if canManage}
					<button class="skip-btn" onclick={skipTrack} disabled={isSkipping} title="Skip Track">
						<SkipForward size={24} color="var(--auxie-cloud-white-50)" />
					</button>
				{/if}
			</div>
		</div>
	{:else}
		<div class="empty-now-playing">
			<p>Waiting for music to start...</p>
		</div>
	{/if}
</div>

<style>
	.now-playing {
		margin-bottom: 20px;
	}

	.section-title {
		font-size: 14px;
		color: var(--auxie-cloud-white-400);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin: 0 0 12px 5px;
	}

	.playing-card {
		display: flex;
		align-items: center;
		gap: 15px;
		background: linear-gradient(135deg, rgba(138, 43, 226, 0.2), rgba(0, 255, 135, 0.2));
		border: 1px solid rgba(138, 43, 226, 0.3);
		border-radius: 16px;
		padding: 15px;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
	}

	.playing-cover {
		width: 60px;
		height: 60px;
		border-radius: 12px;
		object-fit: cover;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
	}

	.playing-info {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.playing-info h4 {
		margin: 0;
		color: var(--auxie-cloud-white-50);
		font-size: 16px;
		font-weight: 600;
	}

	.playing-info p {
		margin: 0;
		color: var(--auxie-cloud-white-400);
		font-size: 14px;
	}

	.status-text {
		display: flex;
		gap: 8px;
		margin-top: 4px !important;
	}

	.status-badge {
		font-size: 12px;
		padding: 4px 8px;
		border-radius: 6px;
		font-weight: 500;
	}

	.status-badge.playing {
		background: rgba(0, 255, 135, 0.2);
		color: #00ff87;
	}

	.status-badge.paused {
		background: rgba(138, 43, 226, 0.2);
		color: #8a2be2;
	}

	.status-badge.idle {
		background: rgba(255, 255, 255, 0.1);
		color: var(--auxie-cloud-white-400);
	}

	.playing-actions {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.playing-indicator {
		display: flex;
		gap: 3px;
		height: 24px;
		align-items: flex-end;
	}

	.bar {
		width: 3px;
		background: var(--auxie-cloud-white-50);
		border-radius: 2px;
		animation: pulse 0.6s ease-in-out infinite;
	}

	.bar:nth-child(1) {
		animation-delay: 0s;
		height: 8px;
	}

	.bar:nth-child(2) {
		animation-delay: 0.15s;
		height: 12px;
	}

	.bar:nth-child(3) {
		animation-delay: 0.3s;
		height: 16px;
	}

	.bar.paused {
		animation: none;
		opacity: 0.5;
	}

	@keyframes pulse {
		0%, 100% {
			height: 8px;
		}
		50% {
			height: 20px;
		}
	}

	.skip-btn {
		background: transparent;
		border: none;
		cursor: pointer;
		color: var(--auxie-cloud-white-50);
		padding: 8px;
		border-radius: 8px;
		transition: all 0.3s;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.skip-btn:hover:not(:disabled) {
		background: rgba(0, 255, 135, 0.2);
		color: #00ff87;
	}

	.skip-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.empty-now-playing {
		padding: 20px;
		text-align: center;
		color: var(--auxie-cloud-white-400);
		font-style: italic;
	}
</style>
