<script lang="ts">
import { onMount } from "svelte";
import SkipForward from "./icons/SkipForward.svelte";
import ThumbsUp from "./icons/ThumbsUp.svelte";
import { globalPlayer } from "$lib/player/player.svelte";
import type { Track } from "$lib/player/types";

let { queue = [], currentUser, slug, ws }: { queue?: any[], currentUser?: any, slug?: string, ws?: WebSocket | null } = $props();

let playbackTrack = $state<any>(null);

let isSkipping = $state(false);
let canManage = $derived(currentUser?.CurrentRole === "Host" || currentUser?.CurrentRole === "DJ");

let likedTracks = $state<Set<number>>(new Set());
let likeCounts = $state<Record<number, number>>({});
let skipVoteCount = $state(0);
let skipVoteThreshold = $state(0);
let hasVotedSkip = $state(false);
let isVotingSkip = $state(false);
let lastTrackId = $state<number | null>(null);

let dominantColor = $state<{r: number, g: number, b: number} | null>(null);

let baseColorStr = $derived(
    dominantColor 
    ? `${dominantColor.r}, ${dominantColor.g}, ${dominantColor.b}` 
    : `138, 43, 226`
);

let progressPercent = $derived(
	globalPlayer.durationMs > 0 
	? (globalPlayer.positionMs / globalPlayer.durationMs) * 100 
	: 0
);

// Sync likeCounts from playback track
$effect(() => {
	if (playbackTrack) {
		const trackId = playbackTrack.room_track_id;
		likeCounts[trackId] = playbackTrack.like_count ?? 0;
	}
});

// Sync playback track from queue prop
$effect(() => {
	const playingTrack = queue.length > 0 && queue[0].status === "playing" ? queue[0] : null;
	if (playingTrack) {
		if (playbackTrack !== playingTrack) {
			playbackTrack = playingTrack;
		}
	} else {
		if (playbackTrack) {
			playbackTrack = null;
		}
	}
});

// Reset skip vote state when track changes
$effect(() => {
	const id = playbackTrack?.room_track_id;
	if (id && id !== lastTrackId) {
		lastTrackId = id;
		skipVoteCount = 0;
		skipVoteThreshold = 0;
		hasVotedSkip = false;
		dominantColor = null;
	} else if (!id) {
		lastTrackId = null;
		dominantColor = null;
	}
});

// Auto-play currently playing track when ready
$effect(() => {
	const track = playbackTrack;
	if (!track) {
		globalPlayer.stop();
		return;
	}
	
	// If the globalPlayer isn't playing this track yet, play it
	if (globalPlayer.currentTrack?.room_track_id !== track.room_track_id) {
		console.log("🚀 [AutoPlay Effect] Starting track:", track.title);
		globalPlayer.play(track as Track).catch(err => {
            console.error("Playback failed", err);
        });
	}
});

async function togglePlayPause() {
	if (!canManage) return;
	await globalPlayer.togglePlay();
}

async function likeTrack(roomTrackId: number) {
	if (!slug) return;
	try {
		const res = await fetch(`/api/v1/room/${slug}/track/${roomTrackId}/like`, { method: 'POST' });
		const data = await res.json();
		if (res.ok) {
			const newSet = new Set(likedTracks);
			if (data.liked) {
				newSet.add(roomTrackId);
			} else {
				newSet.delete(roomTrackId);
			}
			likedTracks = newSet;
			likeCounts = { ...likeCounts, [roomTrackId]: data.like_count };
		}
	} catch(err) {
		console.error(err);
	}
}

async function voteSkip() {
	if (!slug || isVotingSkip || hasVotedSkip) return;
	isVotingSkip = true;
	try {
		const res = await fetch(`/api/v1/room/${slug}/vote-skip`, { method: 'POST' });
		const data = await res.json();
		if (res.ok) {
			hasVotedSkip = true;
			skipVoteCount = data.votes;
			skipVoteThreshold = data.threshold;
		}
	} catch(err) {
		console.error(err);
	} finally {
		isVotingSkip = false;
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
onMount(async () => {
	console.log("🚀 NowPlaying component mounted");
	
	await globalPlayer.init((trackId) => {
		notifyPlaybackEnded(trackId);
	});
});

// Reactive trigger - set listener when ws changes
$effect(() => {
	if (ws) {
		const listener = (event: MessageEvent) => {
			try {
				const msg = JSON.parse(event.data);

				if (msg.type === "playback:start") {
					playbackTrack = msg.track;
					// Reset skip vote status
					skipVoteCount = 0;
					skipVoteThreshold = 0;
					hasVotedSkip = false;
					console.log("🎵 Playback started:", msg.track.title, "Platform:", msg.track.platform);
				} else if (msg.type === "playback:skipped") {
					globalPlayer.stop();
					playbackTrack = null;
					console.log("⏭️  Track skipped");
				} else if (msg.type === "playback:ended") {
					globalPlayer.stop();
					playbackTrack = null;
					console.log("✅ Track ended");
				} else if (msg.type === "TRACK_LIKED") {
					const { room_track_id, like_count } = msg.payload;
					likeCounts = { ...likeCounts, [room_track_id]: like_count };
				} else if (msg.type === "SKIP_VOTE") {
					const { room_track_id, votes, threshold } = msg.payload;
					if (playbackTrack?.room_track_id === room_track_id) {
						skipVoteCount = votes;
						skipVoteThreshold = threshold;
					}
				}
			} catch (err) {
				console.error("Error parsing WS message:", err);
			}
		};

		ws.addEventListener("message", listener);
		return () => {
			ws.removeEventListener("message", listener);
		};
	}
});

// Cleanup on unmount
$effect.pre(() => {
	return () => {
		globalPlayer.stop();
		globalPlayer.disconnect();
	};
});

function handleImageLoad(e: Event) {
	const imgEl = e.target as HTMLImageElement;
	const blockSize = 5;
	const canvas = document.createElement('canvas');
	const context = canvas.getContext && canvas.getContext('2d');
	let data, width, height;
	let i = -4;
	let length;
	let rgb = {r: 0, g: 0, b: 0};
	let count = 0;

	if (!context) return;

	height = canvas.height = imgEl.naturalHeight || imgEl.offsetHeight || imgEl.height;
	width = canvas.width = imgEl.naturalWidth || imgEl.offsetWidth || imgEl.width;

	if (width === 0 || height === 0) return;

	try {
		context.drawImage(imgEl, 0, 0);
		data = context.getImageData(0, 0, width, height);
	} catch(err) {
		console.warn("Could not get image data for color extraction", err);
		return;
	}

	length = data.data.length;

	while ((i += blockSize * 4) < length) {
		++count;
		rgb.r += data.data[i];
		rgb.g += data.data[i+1];
		rgb.b += data.data[i+2];
	}

	if (count > 0) {
		rgb.r = Math.floor(rgb.r / count);
		rgb.g = Math.floor(rgb.g / count);
		rgb.b = Math.floor(rgb.b / count);
		dominantColor = rgb;
	}
}
</script>

<div class="now-playing">
	{#if globalPlayer.error}
		<div class="spotify-error-banner">
			<span class="error-icon">⚠️</span>
			<div class="error-details">
				<strong>Player error</strong>
				<p>{globalPlayer.error}. Please ensure Widevine DRM is enabled in your browser settings if using Spotify, and that you are using localhost/127.0.0.1 or HTTPS.</p>
			</div>
		</div>
	{/if}

	{#if playbackTrack}
		<h3 class="section-title">Now Playing</h3>
		<div class="playing-card" style="background: linear-gradient(90deg, rgba({baseColorStr}, 0.5) {progressPercent}%, rgba({baseColorStr}, 0.1) {progressPercent}%, rgba(0, 255, 135, 0.05) 100%);">
			<img 
				src={playbackTrack.cover_url?.String || playbackTrack.cover_url || "/placeholder.png"} 
				alt={playbackTrack.title} 
				class="playing-cover"
				crossorigin="anonymous"
				onload={handleImageLoad}
			/>
			<div class="playing-info">
				<h4>{playbackTrack.title}</h4>
				<p>{playbackTrack.artist?.String || playbackTrack.artist || "Unknown Artist"}</p>
				<p class="status-text">
					{#if globalPlayer.status === 'playing'}
						<span class="status-badge playing">🎵 Playing</span>
					{:else if globalPlayer.status === 'paused'}
						<span class="status-badge paused">⏸️ Paused</span>
					{:else}
						<span class="status-badge idle">⏹️ Idle</span>
					{/if}
				</p>
			</div>
			<div class="playing-actions">
				<div class="playing-indicator">
					{#if globalPlayer.status === 'playing'}
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
					<button class="play-pause-btn" onclick={togglePlayPause} title={globalPlayer.status === 'playing' ? 'Pause' : 'Play'}>
						{#if globalPlayer.status === 'playing'}
							<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
								<path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
							</svg>
						{:else}
							<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
								<path d="M8 5v14l11-7z"/>
							</svg>
						{/if}
					</button>

					<button class="skip-btn" onclick={skipTrack} disabled={isSkipping} title="Skip Track">
						<SkipForward size={24} color="var(--auxie-cloud-white-50)" />
					</button>

				{/if}
			</div>
		</div>

		{#if canManage}
			<div class="volume-container" title="Volume">
				<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="currentColor" class="vol-icon">
					<path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/>
				</svg>
				<div class="volume-track-wrapper">
					<div class="volume-fill" style="width: {globalPlayer.volume * 100}%; background: linear-gradient(90deg, rgba({baseColorStr}, 0.15), rgba({baseColorStr}, 0.4)); box-shadow: 2px 0 12px rgba({baseColorStr}, 0.3);"></div>
					<span class="volume-label">{Math.round(globalPlayer.volume * 100)}%</span>
					<input 
						type="range" 
						min="0" 
						max="1" 
						step="0.01" 
						value={globalPlayer.volume} 
						oninput={(e) => globalPlayer.setVolume(Number(e.currentTarget.value))}
						class="volume-slider-overlay"
					/>
				</div>
			</div>
		{/if}

		<div class="track-voting">
			<button
				class="vote-btn like-btn {likedTracks.has(playbackTrack.room_track_id) ? 'active' : ''}"
				onclick={() => likeTrack(playbackTrack.room_track_id)}
				title="Like this track"
			>
				<ThumbsUp size={16} color="currentColor" />
				<span>{likeCounts[playbackTrack.room_track_id] ?? playbackTrack.like_count ?? 0}</span>
			</button>

			{#if !canManage}
				<button
					class="vote-btn skip-vote-btn {hasVotedSkip ? 'voted' : ''}"
					onclick={voteSkip}
					disabled={hasVotedSkip || isVotingSkip}
					title={hasVotedSkip ? "You voted to skip" : "Vote to skip this track"}
				>
					<SkipForward size={16} color="currentColor" />
					<span>
						{#if skipVoteThreshold > 0}
							Skip {skipVoteCount}/{skipVoteThreshold}
						{:else}
							Vote skip
						{/if}
					</span>
				</button>
			{/if}
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

	.play-pause-btn {
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

	.skip-btn:hover:not(:disabled), .play-pause-btn:hover:not(:disabled) {
		background: rgba(0, 255, 135, 0.2);
		color: #00ff87;
	}

	.skip-btn:disabled, .play-pause-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.empty-now-playing {
		padding: 20px;
		text-align: center;
		color: var(--auxie-cloud-white-400);
		font-style: italic;
	}

	.spotify-error-banner {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		background: rgba(244, 124, 124, 0.1);
		border: 1px solid rgba(244, 124, 124, 0.35);
		border-radius: 12px;
		padding: 12px 16px;
		margin-bottom: 16px;
		text-align: left;
	}

	.spotify-error-banner .error-icon {
		font-size: 20px;
		line-height: 1;
	}

	.spotify-error-banner .error-details strong {
		color: var(--auxie-soft-crimson-400);
		font-size: 14px;
		display: block;
		margin-bottom: 4px;
	}

	.spotify-error-banner .error-details p {
		margin: 0;
		color: var(--auxie-cloud-white-400);
		font-size: 12px;
		line-height: 1.4;
	}

	/* Voting row */
	.track-voting {
		display: flex;
		gap: 10px;
		margin-top: 12px;
		padding: 0 2px;
	}

	.vote-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 7px 14px;
		border-radius: 20px;
		border: 1px solid rgba(255, 255, 255, 0.1);
		background: rgba(255, 255, 255, 0.04);
		color: var(--auxie-cloud-white-400);
		font-family: "Onest", sans-serif;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.vote-btn:hover:not(:disabled) {
		background: rgba(255, 255, 255, 0.08);
		color: var(--auxie-cloud-white-100);
	}

	.vote-btn.like-btn.active {
		background: rgba(0, 255, 135, 0.12);
		border-color: var(--auxie-intense-mint-500);
		color: var(--auxie-intense-mint-500);
	}

	.vote-btn.skip-vote-btn {
		color: var(--auxie-cloud-white-400);
	}

	.vote-btn.skip-vote-btn.voted {
		background: rgba(138, 43, 226, 0.12);
		border-color: var(--auxie-electric-purple-500);
		color: var(--auxie-electric-purple-400);
		cursor: default;
	}

	.vote-btn:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.volume-container {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-top: 15px;
	}

	.vol-icon {
		color: var(--auxie-cloud-white-400);
		flex-shrink: 0;
	}

	.volume-track-wrapper {
		position: relative;
		flex: 1;
		height: 38px;
		background: linear-gradient(135deg, rgba(255, 255, 255, 0.03), rgba(255, 255, 255, 0.08));
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 12px;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: inset 0 2px 8px rgba(0, 0, 0, 0.2), 0 4px 12px rgba(0, 0, 0, 0.15);
		backdrop-filter: blur(8px);
		transition: border-color 0.2s;
	}

	.volume-track-wrapper:hover {
		border-color: rgba(255, 255, 255, 0.2);
	}

	.volume-fill {
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		pointer-events: none;
		transition: width 0.1s linear;
	}

	.volume-label {
		position: relative;
		z-index: 1;
		font-size: 13px;
		font-weight: 500;
		color: var(--auxie-cloud-white-100);
		pointer-events: none;
		font-family: "Onest", sans-serif;
	}

	.volume-slider-overlay {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		opacity: 0;
		cursor: pointer;
		margin: 0;
	}
</style>
