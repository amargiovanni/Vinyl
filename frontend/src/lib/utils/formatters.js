export function formatDuration(seconds) {
  if (!seconds || seconds <= 0) return '0m';
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

export function formatDurationMS(ms) {
  if (!ms || ms <= 0) return '0:00';
  const totalSec = Math.floor(ms / 1000);
  const min = Math.floor(totalSec / 60);
  const sec = totalSec % 60;
  return `${min}:${sec.toString().padStart(2, '0')}`;
}

export function formatTime(isoString) {
  if (!isoString) return '';
  const d = new Date(isoString);
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false });
}

export function formatTimeRange(start, end) {
  return `${formatTime(start)} — ${formatTime(end)}`;
}

export function formatDate(dateStr) {
  if (!dateStr) return '';
  const d = new Date(dateStr + 'T00:00:00');
  return d.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' });
}

export function formatDateShort(dateStr) {
  if (!dateStr) return '';
  const d = new Date(dateStr + 'T00:00:00');
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

export function isToday(dateStr) {
  return dateStr === new Date().toISOString().split('T')[0];
}

export function isYesterday(dateStr) {
  const yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);
  return dateStr === yesterday.toISOString().split('T')[0];
}

export function weatherEmoji(condition) {
  const map = {
    'Clear': '☀️', 'Clouds': '☁️', 'Rain': '🌧️', 'Drizzle': '🌧️',
    'Thunderstorm': '⛈️', 'Snow': '❄️', 'Mist': '🌫️', 'Fog': '🌫️',
  };
  return map[condition] || '🌤️';
}

export function moodEmoji(mood) {
  const map = {
    happy: '😊', calm: '😌', energetic: '🔥', sad: '😢',
    thoughtful: '🤔', frustrated: '😤', in_love: '🥰', dreamy: '🌙',
    motivated: '💪', celebrating: '🎉', sleepy: '😴', nostalgic: '🌊',
  };
  return map[mood] || '';
}

export function sourceIcon(source) {
  if (source === 'spotify') return '●';
  if (source === 'apple_music') return '●';
  return '●';
}

export function sourceLabel(source) {
  if (source === 'spotify') return 'Spotify';
  if (source === 'apple_music') return 'Apple Music';
  return source;
}
