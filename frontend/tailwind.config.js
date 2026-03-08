/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{svelte,js,ts}'],
  theme: {
    extend: {
      colors: {
        vinyl: {
          black: '#1A1410',
          groove: '#2A2018',
          cardboard: '#3D3024',
          leather: '#5C4A3A',
          wood: '#7A6450',
          amber: '#C49A6C',
          gold: '#E8C496',
          cream: '#F2E6D0',
          white: '#FAF3E8',
        },
        mood: {
          happy: '#D4A96A',
          calm: '#8BA48B',
          energetic: '#C47A5A',
          sad: '#6B7E9E',
          nostalgic: '#9E7A8E',
        },
        heat: {
          0: '#2A2018',
          1: '#5C4A3A',
          2: '#8B6D4F',
          3: '#C49A6C',
          4: '#E8C496',
        },
        source: {
          spotify: '#7BA67B',
          apple: '#B87A8E',
        },
        muted: '#8E7E6E',
        faint: '#6E5E4E',
      },
      fontFamily: {
        display: ['Playfair Display', 'Georgia', 'serif'],
        body: ['Source Serif 4', 'Palatino', 'serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      fontSize: {
        xxs: '0.7rem',
      },
    },
  },
  plugins: [],
}
