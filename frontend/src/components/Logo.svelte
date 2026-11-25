<script lang="ts">
  // Iron Cross mark: silver-edged black cross pattée with curved arms.
  let { size = 32 }: { size?: number } = $props()

  const CROSS =
    'M29 5H71Q63 24 63.5 36.5Q76 37 95 29V71Q76 63 63.5 63.5Q63 76 71 95H29Q37 76 36.5 63.5Q24 63 5 71V29Q24 37 36.5 36.5Q37 24 29 5Z'
</script>

<svg width={size} height={size} viewBox="0 0 100 100" aria-hidden="true">
  <defs>
    <linearGradient id="il-silver" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="#f1ead6" />
      <stop offset="0.5" stop-color="#b9b2a0" />
      <stop offset="1" stop-color="#8d8778" />
    </linearGradient>
  </defs>
  <path d={CROSS} fill="url(#il-silver)" />
  <path d={CROSS} fill="#0d0e0c" transform="translate(50 50) scale(.78) translate(-50 -50)" />
  <path d={CROSS} fill="none" stroke="#d6aa4c" stroke-width="1.6" opacity=".55"
    transform="translate(50 50) scale(.66) translate(-50 -50)" />
</svg>

<style>
  svg {
    display: block;
    flex-shrink: 0;
    filter: drop-shadow(0 1px 2px rgb(0 0 0 / 0.6));
  }
</style>
