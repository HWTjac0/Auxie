<script lang="ts">
import MusicalNote from "./icons/MusicalNote.svelte";
import SkipForward from "./icons/SkipForward.svelte";
import Check from "./icons/Check.svelte";
import Cross from "./icons/Cross.svelte";
import ThumbsUp from "./icons/ThumbsUp.svelte";
import { globalPlayer } from "$lib/player/player.svelte";

let { queue = [], proposedQueue = [], currentUser, slug }: { queue?: any[], proposedQueue?: any[], currentUser?: any, slug?: string } = $props();

let upNext = $derived(queue.length > 0 ? (queue[0].status === 'playing' ? queue.slice(1) : queue) : []);

let canManage = $derived(currentUser?.CurrentRole === "Host" || currentUser?.CurrentRole === "DJ");

let isApproving = $state(false);

let timeRemaining = $derived(
    globalPlayer.durationMs > 0 && globalPlayer.durationMs >= globalPlayer.positionMs
    ? Math.floor((globalPlayer.durationMs - globalPlayer.positionMs) / 1000)
    : 0
);

let formattedTime = $derived(
    timeRemaining > 0
    ? `${Math.floor(timeRemaining / 60)}:${(timeRemaining % 60).toString().padStart(2, '0')}`
    : ""
);

// Track liked state per roomTrackId
let likedTracks = $state<Set<number>>(new Set());
let likeCounts = $state<Record<number, number>>({});

// Sync likeCounts from queue
$effect(() => {
    const counts: Record<number, number> = {};
    for (const t of queue) {
        counts[t.room_track_id] = t.like_count ?? 0;
    }
    likeCounts = counts;
});

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

// Called from parent when WS events arrive
export function onWsMessage(msg: any) {
    if (msg.type === "TRACK_LIKED") {
        const { room_track_id, liked } = msg.payload;
        const newSet = new Set(likedTracks);
        if (liked) {
            likeCounts = { ...likeCounts, [room_track_id]: (likeCounts[room_track_id] ?? 0) + 1 };
        } else {
            likeCounts = { ...likeCounts, [room_track_id]: Math.max(0, (likeCounts[room_track_id] ?? 1) - 1) };
        }
        likedTracks = newSet;
    }
}

async function handleProposed(trackId: number, action: 'approve' | 'reject') {
    if (!slug || isApproving || !canManage) return;
    isApproving = true;
    try {
        const res = await fetch(`/api/v1/room/${slug}/proposed/${trackId}/${action}`, { method: 'POST' });
        if (!res.ok) console.error(`Failed to ${action} track`);
    } catch(err) {
        console.error(err);
    } finally {
        isApproving = false;
    }
}
</script>

<div class="queue-tab">
  {#if canManage && proposedQueue.length > 0}
    <div class="proposed-queue">
      <h3 class="section-title onest-500">Proposed by Guests ({proposedQueue.length})</h3>
      <div class="next-list">
        {#each proposedQueue as track (track.room_track_id)}
          <div class="next-item proposed-item">
            <img src={track.cover_url?.String || track.cover_url || "/placeholder.png"} alt={track.title} class="next-cover" />
            <div class="next-info">
              <h4 class="onest-500">{track.title}</h4>
              <p class="onest-300">{track.artist?.String || track.artist || "Unknown Artist"}</p>
            </div>
            <div class="proposed-actions">
              <button class="action-btn approve" onclick={() => handleProposed(track.room_track_id, 'approve')} disabled={isApproving}>
                <Check size={20} color="currentColor" />
              </button>
              <button class="action-btn reject" onclick={() => handleProposed(track.room_track_id, 'reject')} disabled={isApproving}>
                <Cross size={20} color="currentColor" />
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <div class="queue-header">
    <h2 class="onest-500">
        Queue
        {#if formattedTime}
            <span class="time-remaining">({formattedTime} left)</span>
        {/if}
    </h2>
  </div>
  
  {#if queue.length === 0 || (queue.length === 1 && queue[0].status === "playing")}
    <div class="empty-state">
      <div class="empty-icon">
        <MusicalNote size={40}/>
      </div>
      <h3 class="onest-500">Queue is empty</h3>
      <p class="onest-300">Search for desired track to be played below and add it to the queue!</p>
    </div>
  {:else}
    {#if upNext.length > 0}
      <div class="up-next">
        <h3 class="section-title onest-500">Up Next</h3>
        <div class="next-list">
          {#each upNext as track (track.room_track_id)}
            <div class="next-item">
              <img src={track.cover_url?.String || track.cover_url || "/placeholder.png"} alt={track.title} class="next-cover" />
              <div class="next-info">
                <h4 class="onest-500">{track.title}</h4>
                <p class="onest-300">{track.artist?.String || track.artist || "Unknown Artist"}</p>
              </div>
              <button
                class="like-small {likedTracks.has(track.room_track_id) ? 'active' : ''}"
                onclick={() => likeTrack(track.room_track_id)}
                title="Like"
              >
                <ThumbsUp size={14} color="currentColor" />
                <span>{likeCounts[track.room_track_id] ?? track.like_count ?? 0}</span>
              </button>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .queue-tab {
    display: flex;
    flex-direction: column;
    gap: 15px;
    padding: 10px 0;
  }

  .queue-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 5px;
  }

  h2 {
    font-size: 18px;
    color: var(--auxie-cloud-white-50);
    margin: 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .time-remaining {
    font-size: 13px;
    color: var(--auxie-cloud-white-400);
    font-weight: 400;
  }

  .section-title {
    font-size: 14px;
    color: var(--auxie-cloud-white-400);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 12px;
    padding-left: 5px;
  }

  /* Currently Playing */
  .currently-playing {
    margin-bottom: 10px;
  }

  .playing-card {
    display: flex;
    align-items: center;
    gap: 15px;
    background: linear-gradient(135deg, rgba(138, 43, 226, 0.15), rgba(0, 255, 135, 0.15));
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 16px;
    padding: 15px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
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
  }

  .playing-info p {
    margin: 0;
    color: var(--auxie-cloud-white-400);
    font-size: 14px;
  }

  .playing-actions {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 15px;
    padding-right: 10px;
  }

  .skip-btn {
    background: rgba(255, 255, 255, 0.1);
    border: none;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .skip-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.2);
    transform: scale(1.05);
  }

  .skip-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Equalizer Animation */
  .playing-indicator {
    display: flex;
    align-items: flex-end;
    gap: 3px;
    height: 20px;
  }

  .playing-indicator .bar {
    width: 4px;
    background-color: var(--auxie-intense-mint-500);
    border-radius: 2px;
    animation: equalize 1s infinite alternate;
  }

  .playing-indicator .bar:nth-child(1) { height: 10px; animation-delay: -0.2s; }
  .playing-indicator .bar:nth-child(2) { height: 20px; animation-delay: -0.4s; }
  .playing-indicator .bar:nth-child(3) { height: 15px; animation-delay: -0.6s; }

  @keyframes equalize {
    0% { height: 4px; }
    100% { height: 20px; }
  }

  /* Up Next */
  .up-next {
    display: flex;
    flex-direction: column;
  }

  .next-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .next-item {
    display: flex;
    align-items: center;
    gap: 12px;
    background-color: rgba(255, 255, 255, 0.03);
    border-radius: 12px;
    padding: 10px;
    transition: background-color 0.2s;
  }

  .next-item:hover {
    background-color: rgba(255, 255, 255, 0.06);
  }

  .next-cover {
    width: 44px;
    height: 44px;
    border-radius: 8px;
    object-fit: cover;
  }

  .next-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .next-info h4 {
    margin: 0;
    color: var(--auxie-cloud-white-100);
    font-size: 15px;
  }

  .next-info p {
    margin: 0;
    color: var(--auxie-cloud-white-500);
    font-size: 13px;
  }

  /* Empty State */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    text-align: center;
    padding: 40px 20px;
    background-color: var(--auxie-deep-navy-700);
    border: 1px dashed var(--auxie-deep-navy-500);
    border-radius: 16px;
  }

  .empty-icon {
    margin-bottom: 10px;
    opacity: 0.5;
    filter: drop-shadow(0 0 10px rgba(255, 255, 255, 0.2));
  }

  .empty-state h3 {
    color: var(--auxie-cloud-white-200);
    margin: 0;
    font-size: 18px;
  }

  .empty-state p {
    color: var(--auxie-cloud-white-600);
    margin: 0;
    font-size: 14px;
    max-width: 250px;
    line-height: 1.4;
  }

  .proposed-queue {
    margin-bottom: 25px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 15px;
    padding: 15px;
  }

  .proposed-item {
    background: rgba(255, 255, 255, 0.05);
    border-radius: 10px;
  }

  .proposed-actions {
    display: flex;
    gap: 8px;
    margin-left: auto;
  }

  .action-btn {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-weight: bold;
    color: white;
    transition: all 0.2s;
  }

  .action-btn.approve {
    background: rgba(46, 204, 113, 0.2);
    color: #2ecc71;
  }

  .action-btn.approve:hover:not(:disabled) {
    background: rgba(46, 204, 113, 0.4);
    transform: scale(1.1);
  }

  .action-btn.reject {
    background: rgba(231, 76, 60, 0.2);
    color: #e74c3c;
  }

  .action-btn.reject:hover:not(:disabled) {
    background: rgba(231, 76, 60, 0.4);
    transform: scale(1.1);
  }

  .action-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Voting row */
  .track-voting {
    display: flex;
    gap: 10px;
    margin-top: 10px;
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

  /* Like button on queue items */
  .like-small {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    border-radius: 14px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    background: transparent;
    color: var(--auxie-cloud-white-500);
    font-family: "Onest", sans-serif;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
    flex-shrink: 0;
  }

  .like-small:hover {
    background: rgba(0, 255, 135, 0.08);
    color: var(--auxie-intense-mint-400);
  }

  .like-small.active {
    background: rgba(0, 255, 135, 0.12);
    border-color: var(--auxie-intense-mint-500);
    color: var(--auxie-intense-mint-500);
  }
</style>
