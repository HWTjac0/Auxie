<script lang="ts">
import TextInput from "./TextInput.svelte";

let { searchQuery = $bindable("") }: { searchQuery?: string } = $props();

let dialogElement: HTMLDialogElement;
let spotifyTracks = $state<any[]>([]);
let tidalTracks = $state<any[]>([]);
let loading = $state(false);

export function show() {
  dialogElement.showModal();
}

export function close() {
  dialogElement.close();
}

function formatDuration(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function parseTidalResults(json: any): any[] {
  if (!json || !json.included) return [];
  const included = json.included;

  const artistsMap = new Map<string, string>();
  const albumsMap = new Map<string, { name: string; coverUrl: string }>();

  for (const item of included) {
    if (item.type === "artists") {
      artistsMap.set(item.id, item.attributes?.name || "");
    } else if (item.type === "albums") {
      const name = item.attributes?.title || "";
      // Tidal V2 coverArt relation is to-one, so data is a single object
      const coverArtId = item.relationships?.coverArt?.data?.id;
      let coverUrl = "";
      if (coverArtId) {
        const parts = coverArtId.replace(/-/g, "/");
        coverUrl = `https://resources.tidal.com/images/${parts}/160x160.jpg`;
      }
      albumsMap.set(item.id, { name, coverUrl });
    }
  }

  const tracksList: any[] = [];

  for (const item of included) {
    if (item.type === "tracks") {
      const id = item.id;
      const title = item.attributes?.title || "";
      const version = item.attributes?.version;
      const name = version ? `${title} (${version})` : title;

      // Parse ISO 8601 duration "PT4M41S"
      const durationStr = item.attributes?.duration || "";
      const durationMatches = durationStr.match(
        /^PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$/,
      );
      let duration_ms = 0;
      if (durationMatches) {
        const hours = parseInt(durationMatches[1] || "0", 10);
        const minutes = parseInt(durationMatches[2] || "0", 10);
        const seconds = parseInt(durationMatches[3] || "0", 10);
        duration_ms = (hours * 3600 + minutes * 60 + seconds) * 1000;
      }

      // Resolve artists
      const artistRelations = item.relationships?.artists?.data || [];
      const artists = artistRelations
        .map((ref: any) => {
          const artistName = artistsMap.get(ref.id);
          return artistName ? { name: artistName } : null;
        })
        .filter(Boolean);

      // Resolve album (Tidal V2 album relation is to-one)
      const albumId = item.relationships?.album?.data?.id;
      const albumInfo = albumId ? albumsMap.get(albumId) : null;
      const albumName = albumInfo?.name || "";
      const coverUrl = albumInfo?.coverUrl || "";

      tracksList.push({
        id,
        name,
        artists,
        album: {
          name: albumName,
          images: coverUrl ? [{ url: coverUrl }] : [],
        },
        duration_ms,
        platform: "tidal",
        uri: `tidal:track:${id}`,
      });
    }
  }

  return tracksList;
}

$effect(() => {
  const query = searchQuery.trim();
  if (!query) {
    spotifyTracks = [];
    tidalTracks = [];
    loading = false;
    return;
  }

  loading = true;
  const timeoutId = setTimeout(() => {
    fetch(`/api/v1/search?search=${encodeURIComponent(query)}`)
      .then((res) => res.json())
      .then((json) => {
        spotifyTracks = json.spotify?.tracks?.items || [];
        tidalTracks = parseTidalResults(json.tidal) || [];
        loading = false;
      })
      .catch((err) => {
        console.error(err);
        loading = false;
      });
  }, 300);

  return () => {
    clearTimeout(timeoutId);
  };
});
</script>

<dialog bind:this={dialogElement} class="search-dialog">
  <div class="dialog-content">
    <div class="dialog-header">
      <h2 class="onest-500">Search for a song</h2>
      <button class="close-btn" onclick={close}>&times;</button>
    </div>
    
    <div class="search-section">
      <TextInput bind:value={searchQuery} placeholder="Type song name..." />
      <div class="search-results" class:centered-state={loading || (spotifyTracks.length === 0 && tidalTracks.length === 0)}>
        {#if loading}
          <div class="spinner"></div>
        {:else if spotifyTracks.length > 0 || tidalTracks.length > 0}
          <div class="results-container">
            {#if spotifyTracks.length > 0}
              <div class="service-section">
                <h3 class="service-title onest-600 spotify-color">Spotify</h3>
                <div class="tracks-list">
                  {#each spotifyTracks as track (track.id)}
                    <div class="track-item">
                      {#if track.album?.images?.length > 0}
                        <img class="cover-art" src={track.album.images[track.album.images.length - 1].url} alt={track.album.name} />
                      {:else}
                        <div class="cover-placeholder"></div>
                      {/if}
                      <div class="track-info">
                        <div class="track-title onest-500">{track.name}</div>
                        <div class="track-meta onest-400">
                          <span class="artists">{track.artists.map((a: any) => a.name).join(', ')}</span>
                          <span class="bullet">&bull;</span>
                          <span class="album">{track.album.name}</span>
                        </div>
                      </div>
                      <div class="track-duration onest-400">{formatDuration(track.duration_ms)}</div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}

            {#if tidalTracks.length > 0}
              <div class="service-section">
                <h3 class="service-title onest-600 tidal-color">Tidal</h3>
                <div class="tracks-list">
                  {#each tidalTracks as track (track.id)}
                    <div class="track-item">
                      {#if track.album?.images?.length > 0}
                        <img class="cover-art" src={track.album.images[track.album.images.length - 1].url} alt={track.album.name} />
                      {:else}
                        <div class="cover-placeholder"></div>
                      {/if}
                      <div class="track-info">
                        <div class="track-title onest-500">{track.name}</div>
                        <div class="track-meta onest-400">
                          <span class="artists">{track.artists.map((a: any) => a.name).join(', ')}</span>
                          <span class="bullet">&bull;</span>
                          <span class="album">{track.album.name}</span>
                        </div>
                      </div>
                      <div class="track-duration onest-400">{formatDuration(track.duration_ms)}</div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {:else}
          <p class="empty-state">No matching songs found</p>
        {/if}
      </div>
    </div>
  </div>
</dialog>

<style>
  .search-dialog {
    padding: 0;
    border: none;
    border-radius: 20px;
    background-color: var(--auxie-deep-navy-700);
    color: var(--auxie-cloud-white-50);
    max-width: 500px;
    width: 90%;
    margin: auto;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
  }

  .search-dialog::backdrop {
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
  }

  .dialog-content {
    display: flex;
    flex-direction: column;
    padding: 20px;
    gap: 20px;
  }

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .dialog-header h2 {
    margin: 0;
    font-size: 20px;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--auxie-cloud-white-400);
    font-size: 28px;
    cursor: pointer;
    line-height: 1;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.2s;
  }

  .close-btn:hover {
    color: var(--auxie-soft-crimson-400);
  }

  .search-section {
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  .search-results {
    background-color: var(--auxie-deep-navy-600);
    border-radius: 12px;
    padding: 10px;
    height: 350px;
    max-height: 350px;
    overflow-y: auto;
    border: 1px solid var(--auxie-deep-navy-500);
  }

  .centered-state {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .empty-state {
    color: var(--auxie-cloud-white-600);
    font-family: "Onest", sans-serif;
  }

  .tracks-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
  }

  .track-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 12px;
    border-radius: 12px;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .track-item:hover {
    background-color: var(--auxie-deep-navy-500);
  }

  .cover-art {
    width: 44px;
    height: 44px;
    border-radius: 8px;
    object-fit: cover;
    background-color: var(--auxie-deep-navy-800);
  }

  .cover-placeholder {
    width: 44px;
    height: 44px;
    border-radius: 8px;
    background-color: var(--auxie-deep-navy-800);
  }

  .track-info {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }

  .track-title {
    font-size: 14px;
    color: var(--auxie-cloud-white-100);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .track-meta {
    font-size: 12px;
    color: var(--auxie-cloud-white-400);
    display: flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .artists {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 150px;
  }

  .album {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 150px;
  }

  .bullet {
    color: var(--auxie-cloud-white-600);
  }

  .track-duration {
    font-size: 13px;
    color: var(--auxie-cloud-white-400);
    padding-left: 8px;
  }

  /* Spinner */
  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--auxie-deep-navy-500);
    border-top-color: var(--auxie-electric-purple-500);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .results-container {
    display: flex;
    flex-direction: column;
    gap: 20px;
    width: 100%;
  }

  .service-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }

  .service-title {
    font-size: 14px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: 0 0 4px 4px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .spotify-color {
    color: #1DB954;
  }

  .tidal-color {
    color: #00E6FF;
  }
</style>
